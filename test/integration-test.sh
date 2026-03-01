#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
BINARY="$REPO_ROOT/gitbackup"

if [[ -t 1 ]]; then
    GREEN='\033[0;32m'
    RED='\033[0;31m'
    BOLD='\033[1m'
    RESET='\033[0m'
else
    GREEN='' RED='' BOLD='' RESET=''
fi

# Load env vars
if [[ -f "$SCRIPT_DIR/.env" ]]; then
    set -a
    source "$SCRIPT_DIR/.env"
    set +a
fi

VERBOSE=false
TEST_SSH=false
TEST_HTTPS=false
SERVICES=()

for arg in "$@"; do
    case "$arg" in
        -v|--verbose) VERBOSE=true ;;
        --ssh|-ssh) TEST_SSH=true ;;
        --https|-https) TEST_HTTPS=true ;;
        *) SERVICES+=("$arg") ;;
    esac
done

# If neither is specified, test both
if ! $TEST_SSH && ! $TEST_HTTPS; then
    TEST_SSH=true
    TEST_HTTPS=true
fi

if [[ ${#SERVICES[@]} -eq 0 ]]; then
    SERVICES=(github gitlab bitbucket forgejo)
fi

# Expected repo names (same across all services)
EXPECTED_REPOS="gitbackup-test-public gitbackup-test-private"

log() {
    local ts
    ts=$(date +%H:%M:%S)
    echo -e "[$ts] $*"
}

START_TIME=$SECONDS

# --- Helpers ---

pass() {
    log "  ${GREEN}PASS${RESET}: $1"
    RESULTS+=("${GREEN}PASS${RESET}: $1")
    PASSED=$((PASSED + 1))
}

fail() {
    log "  ${RED}FAIL${RESET}: $1"
    RESULTS+=("${RED}FAIL${RESET}: $1")
    FAILED=$((FAILED + 1))
}

check_env() {
    local service="$1"
    shift
    for var in "$@"; do
        if [[ -z "${!var:-}" ]]; then
            log "Skipping $service: $var is not set"
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

    log ""
    log "${BOLD}=== $service ($label) ===${RESET}"

    # Test 1: Fresh clone
    tmpdir=$(mktemp -d)
    trap "rm -rf $tmpdir" RETURN

    log "  Running fresh clone..."
    if run_gitbackup -service "$service" -backupdir "$tmpdir" $extra_flags; then
        all_found=true
        for repo_name in $EXPECTED_REPOS; do
            if check_repo_exists "$tmpdir" "$repo_name"; then
                log "    Found $repo_name"
            else
                log "    Missing $repo_name"
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
    log "  Running update..."
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

    log "  Running bare clone..."
    if run_gitbackup -service "$service" -backupdir "$tmpdir" -bare $extra_flags; then
        all_found=true
        for repo_name in $EXPECTED_REPOS; do
            if check_bare_repo_exists "$tmpdir" "$repo_name"; then
                log "    Found ${repo_name}.git (bare)"
            else
                log "    Missing ${repo_name}.git (bare)"
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

    log "  Running ignore-private clone..."
    if run_gitbackup -service "$service" -backupdir "$tmpdir" -ignore-private $extra_flags; then
        if check_repo_exists "$tmpdir" "gitbackup-test-public"; then
            log "    Found gitbackup-test-public"
        else
            log "    Missing gitbackup-test-public"
            fail "$service ($label): ignore-private — public repo not found"
            rm -rf "$tmpdir"
            trap - RETURN
            return
        fi
        if check_repo_exists "$tmpdir" "gitbackup-test-private"; then
            log "    Found gitbackup-test-private (unexpected)"
            fail "$service ($label): ignore-private — private repo should have been skipped"
        else
            log "    Correctly skipped gitbackup-test-private"
            pass "$service ($label): ignore-private"
        fi
    else
        fail "$service ($label): ignore-private — gitbackup exited with error"
    fi

    rm -rf "$tmpdir"

    # Test 5: Ignore fork
    tmpdir=$(mktemp -d)

    log "  Running clone without -ignore-fork (fork should be present)..."
    if run_gitbackup -service "$service" -backupdir "$tmpdir" $extra_flags; then
        if check_repo_exists "$tmpdir" "gitbackup-test-fork"; then
            log "    Found gitbackup-test-fork (forked repo)"
            pass "$service ($label): fork present without -ignore-fork"
        else
            log "    Missing gitbackup-test-fork (forked repo)"
            fail "$service ($label): fork present without -ignore-fork — gitbackup-test-fork not found"
        fi
    else
        fail "$service ($label): fork present without -ignore-fork — gitbackup exited with error"
    fi

    rm -rf "$tmpdir"

    # Test 6: Ignore fork (with flag)
    tmpdir=$(mktemp -d)

    log "  Running clone with -ignore-fork..."
    if run_gitbackup -service "$service" -backupdir "$tmpdir" -ignore-fork $extra_flags; then
        if check_repo_exists "$tmpdir" "gitbackup-test-fork"; then
            log "    Found gitbackup-test-fork (unexpected — should be skipped)"
            fail "$service ($label): ignore-fork — forked repo should have been skipped"
        else
            log "    Correctly skipped gitbackup-test-fork"
            # Verify non-fork repos are still present
            all_found=true
            for repo_name in $EXPECTED_REPOS; do
                if check_repo_exists "$tmpdir" "$repo_name"; then
                    log "    Found $repo_name"
                else
                    log "    Missing $repo_name"
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

run_starred_tests() {
    local service="$1"
    local label="$2"
    local extra_flags="$3"
    local tmpdir

    log ""
    log "${BOLD}=== $service ($label) ===${RESET}"

    tmpdir=$(mktemp -d)
    trap "rm -rf $tmpdir" RETURN

    log "  Running starred repos clone..."
    if run_gitbackup -service "$service" -backupdir "$tmpdir" -use-https-clone $extra_flags; then
        if check_repo_exists "$tmpdir" "gitbackup-test-starred"; then
            log "    Found gitbackup-test-starred"
        else
            log "    Missing gitbackup-test-starred"
            fail "$service ($label): starred — gitbackup-test-starred not found"
            rm -rf "$tmpdir"
            trap - RETURN
            return
        fi
        local unexpected_found=false
        for repo_name in gitbackup-test-public gitbackup-test-private; do
            if check_repo_exists "$tmpdir" "$repo_name"; then
                log "    Found $repo_name (unexpected — not starred)"
                unexpected_found=true
            else
                log "    Correctly excluded $repo_name (not starred)"
            fi
        done
        if $unexpected_found; then
            fail "$service ($label): starred — non-starred repos should not be present"
        else
            pass "$service ($label): starred"
        fi
    else
        fail "$service ($label): starred — gitbackup exited with error"
    fi

    rm -rf "$tmpdir"
    trap - RETURN
}

# --- Main ---

PASSED=0
FAILED=0
RESULTS=()

log "Building gitbackup..."
(cd "$REPO_ROOT" && go build -o "$BINARY" .)

for service in "${SERVICES[@]}"; do
    case "$service" in
        github)
            check_env github GITHUB_TOKEN || continue
            $TEST_SSH && run_service_tests github "SSH"
            $TEST_HTTPS && run_service_tests github "HTTPS" "-use-https-clone"
            $TEST_HTTPS && run_starred_tests github "starred" "-github.repoType starred"
            ;;
        gitlab)
            check_env gitlab GITLAB_TOKEN || continue
            $TEST_SSH && run_service_tests gitlab "SSH" "-gitlab.projectVisibility all -gitlab.projectMembershipType owner"
            $TEST_HTTPS && run_service_tests gitlab "HTTPS" "-gitlab.projectVisibility all -gitlab.projectMembershipType owner -use-https-clone"
            $TEST_HTTPS && run_starred_tests gitlab "starred" "-gitlab.projectVisibility all -gitlab.projectMembershipType starred"
            ;;
        bitbucket)
            check_env bitbucket BITBUCKET_USERNAME BITBUCKET_TOKEN || continue
            $TEST_SSH && run_service_tests bitbucket "SSH"
            $TEST_HTTPS && run_service_tests bitbucket "HTTPS" "-use-https-clone"
            ;;
        forgejo)
            check_env forgejo FORGEJO_TOKEN || continue
            $TEST_SSH && run_service_tests forgejo "SSH" "-githost.url https://codeberg.org"
            $TEST_HTTPS && run_service_tests forgejo "HTTPS" "-githost.url https://codeberg.org -use-https-clone"
            $TEST_HTTPS && run_starred_tests forgejo "starred" "-githost.url https://codeberg.org -forgejo.repoType starred"
            ;;
    esac
done

# --- Summary (verbose only) ---

log ""
log "${BOLD}==============================${RESET}"
log "Results: $PASSED passed, $FAILED failed"
log ""
for r in "${RESULTS[@]}"; do
    log "  $r"
done
log "Elapsed: $((SECONDS - START_TIME))s"
log "${BOLD}==============================${RESET}"

if [[ $FAILED -gt 0 ]]; then
    exit 1
fi
