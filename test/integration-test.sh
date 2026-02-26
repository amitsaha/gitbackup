#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
BINARY="$REPO_ROOT/gitbackup"

# Load env vars
if [[ -f "$SCRIPT_DIR/.env" ]]; then
    set -a
    source "$SCRIPT_DIR/.env"
    set +a
fi

VERBOSE=false
SERVICES=()

for arg in "$@"; do
    case "$arg" in
        -v|--verbose) VERBOSE=true ;;
        *) SERVICES+=("$arg") ;;
    esac
done

if [[ ${#SERVICES[@]} -eq 0 ]]; then
    SERVICES=(github gitlab bitbucket forgejo)
fi

# Expected repo names (same across all services)
EXPECTED_REPOS="gitbackup-test-public gitbackup-test-private"

# --- Helpers ---

pass() {
    echo "  PASS: $1"
    RESULTS+=("PASS: $1")
    PASSED=$((PASSED + 1))
}

fail() {
    echo "  FAIL: $1"
    RESULTS+=("FAIL: $1")
    FAILED=$((FAILED + 1))
}

check_env() {
    local service="$1"
    shift
    for var in "$@"; do
        if [[ -z "${!var:-}" ]]; then
            echo "Skipping $service: $var is not set"
            return 1
        fi
    done
    return 0
}

is_git_repo() {
    git -C "$1" rev-parse 2>/dev/null
}

is_bare_repo() {
    [[ "$(git -C "$1" rev-parse --is-bare-repository 2>/dev/null)" == "true" ]]
}

# Check that a repo with the given name exists under the backup dir and is a valid git repo
check_repo_exists() {
    local backup_dir="$1"
    local repo_name="$2"
    local dir
    dir=$(find "$backup_dir" -type d -name "$repo_name" -maxdepth 4 2>/dev/null | head -1)
    [[ -n "$dir" ]] && is_git_repo "$dir"
}

# Check that a bare repo with the given name exists under the backup dir
check_bare_repo_exists() {
    local backup_dir="$1"
    local repo_name="$2"
    local dir
    dir=$(find "$backup_dir" -type d -name "${repo_name}.git" -maxdepth 4 2>/dev/null | head -1)
    [[ -n "$dir" ]] && is_bare_repo "$dir"
}

run_gitbackup() {
    if $VERBOSE; then
        $BINARY "$@" 2>&1
    else
        $BINARY "$@" >/dev/null 2>&1
    fi
}

run_service_tests() {
    local service="$1"
    local label="$2"
    local extra_flags="${3:-}"
    local tmpdir

    echo ""
    echo "=== $service ($label) ==="

    # Test 1: Fresh clone
    tmpdir=$(mktemp -d)
    trap "rm -rf $tmpdir" RETURN

    echo "  Running fresh clone..."
    if run_gitbackup -service "$service" -backupdir "$tmpdir" $extra_flags; then
        all_found=true
        for repo_name in $EXPECTED_REPOS; do
            if check_repo_exists "$tmpdir" "$repo_name"; then
                echo "    Found $repo_name"
            else
                echo "    Missing $repo_name"
                all_found=false
            fi
        done
        if $all_found; then
            pass "$service ($label): fresh clone"
        else
            fail "$service ($label): fresh clone — expected repos not found"
        fi
    else
        fail "$service ($label): fresh clone — gitbackup exited with error"
    fi

    # Test 2: Update (run again into same directory)
    echo "  Running update..."
    if run_gitbackup -service "$service" -backupdir "$tmpdir" $extra_flags; then
        all_found=true
        for repo_name in $EXPECTED_REPOS; do
            if ! check_repo_exists "$tmpdir" "$repo_name"; then
                all_found=false
            fi
        done
        if $all_found; then
            pass "$service ($label): update"
        else
            fail "$service ($label): update — repos missing after update"
        fi
    else
        fail "$service ($label): update — gitbackup exited with error"
    fi

    rm -rf "$tmpdir"

    # Test 3: Bare clone
    tmpdir=$(mktemp -d)

    echo "  Running bare clone..."
    if run_gitbackup -service "$service" -backupdir "$tmpdir" -bare $extra_flags; then
        all_found=true
        for repo_name in $EXPECTED_REPOS; do
            if check_bare_repo_exists "$tmpdir" "$repo_name"; then
                echo "    Found ${repo_name}.git (bare)"
            else
                echo "    Missing ${repo_name}.git (bare)"
                all_found=false
            fi
        done
        if $all_found; then
            pass "$service ($label): bare clone"
        else
            fail "$service ($label): bare clone — expected bare repos not found"
        fi
    else
        fail "$service ($label): bare clone — gitbackup exited with error"
    fi

    rm -rf "$tmpdir"

    # Test 4: Ignore private
    tmpdir=$(mktemp -d)

    echo "  Running ignore-private clone..."
    if run_gitbackup -service "$service" -backupdir "$tmpdir" -ignore-private $extra_flags; then
        if check_repo_exists "$tmpdir" "gitbackup-test-public"; then
            echo "    Found gitbackup-test-public"
        else
            echo "    Missing gitbackup-test-public"
            fail "$service ($label): ignore-private — public repo not found"
            rm -rf "$tmpdir"
            trap - RETURN
            return
        fi
        if check_repo_exists "$tmpdir" "gitbackup-test-private"; then
            echo "    Found gitbackup-test-private (unexpected)"
            fail "$service ($label): ignore-private — private repo should have been skipped"
        else
            echo "    Correctly skipped gitbackup-test-private"
            pass "$service ($label): ignore-private"
        fi
    else
        fail "$service ($label): ignore-private — gitbackup exited with error"
    fi

    rm -rf "$tmpdir"

    # Test 5: Ignore fork
    tmpdir=$(mktemp -d)

    echo "  Running clone without -ignore-fork (fork should be present)..."
    if run_gitbackup -service "$service" -backupdir "$tmpdir" $extra_flags; then
        if check_repo_exists "$tmpdir" "hello-world"; then
            echo "    Found hello-world (forked repo)"
            pass "$service ($label): fork present without -ignore-fork"
        else
            echo "    Missing hello-world (forked repo)"
            fail "$service ($label): fork present without -ignore-fork — hello-world not found"
        fi
    else
        fail "$service ($label): fork present without -ignore-fork — gitbackup exited with error"
    fi

    rm -rf "$tmpdir"

    # Test 6: Ignore fork (with flag)
    tmpdir=$(mktemp -d)

    echo "  Running clone with -ignore-fork..."
    if run_gitbackup -service "$service" -backupdir "$tmpdir" -ignore-fork $extra_flags; then
        if check_repo_exists "$tmpdir" "hello-world"; then
            echo "    Found hello-world (unexpected — should be skipped)"
            fail "$service ($label): ignore-fork — forked repo should have been skipped"
        else
            echo "    Correctly skipped hello-world"
            # Verify non-fork repos are still present
            all_found=true
            for repo_name in $EXPECTED_REPOS; do
                if check_repo_exists "$tmpdir" "$repo_name"; then
                    echo "    Found $repo_name"
                else
                    echo "    Missing $repo_name"
                    all_found=false
                fi
            done
            if $all_found; then
                pass "$service ($label): ignore-fork"
            else
                fail "$service ($label): ignore-fork — non-fork repos missing"
            fi
        fi
    else
        fail "$service ($label): ignore-fork — gitbackup exited with error"
    fi

    rm -rf "$tmpdir"
    trap - RETURN
}

# --- Main ---

PASSED=0
FAILED=0
RESULTS=()

echo "Building gitbackup..."
(cd "$REPO_ROOT" && go build -o "$BINARY" .)

for service in "${SERVICES[@]}"; do
    case "$service" in
        github)
            check_env github GITHUB_TOKEN || continue
            run_service_tests github "SSH"
            run_service_tests github "HTTPS" "-use-https-clone"
            ;;
        gitlab)
            check_env gitlab GITLAB_TOKEN || continue
            run_service_tests gitlab "SSH" "-gitlab.projectVisibility all -gitlab.projectMembershipType owner"
            run_service_tests gitlab "HTTPS" "-gitlab.projectVisibility all -gitlab.projectMembershipType owner -use-https-clone"
            ;;
        bitbucket)
            check_env bitbucket BITBUCKET_USERNAME BITBUCKET_TOKEN || continue
            run_service_tests bitbucket "SSH"
            run_service_tests bitbucket "HTTPS" "-use-https-clone"
            ;;
        forgejo)
            check_env forgejo FORGEJO_TOKEN || continue
            run_service_tests forgejo "SSH" "-githost.url https://codeberg.org"
            run_service_tests forgejo "HTTPS" "-githost.url https://codeberg.org -use-https-clone"
            ;;
    esac
done

# --- Summary (verbose only) ---

if $VERBOSE; then
    echo ""
    echo "=============================="
    echo "Results: $PASSED passed, $FAILED failed"
    echo ""
    for r in "${RESULTS[@]}"; do
        echo "  $r"
    done
    echo "=============================="
fi

if [[ $FAILED -gt 0 ]]; then
    exit 1
fi
