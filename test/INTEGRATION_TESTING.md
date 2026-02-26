# Integration Testing

Manual integration test script for verifying gitbackup works end-to-end against real git hosting services before releases.

## Prerequisites

- Go toolchain
- Git
- SSH keys configured for the services you want to test (unless testing HTTPS-only)

## Test Account Setup

Each contributor creates their own test accounts. The script expects a standard set of repos on each service.

### GitHub

1. Create a [personal access token](https://github.com/settings/tokens) with `repo` scope
2. Create two repositories:
   - `gitbackup-test-public` (public)
   - `gitbackup-test-private` (private)
3. Fork any public repo and rename it to `gitbackup-test-ignore-fork` (for testing `-ignore-fork`)

### GitLab

1. Create a [personal access token](https://gitlab.com/-/user_settings/personal_access_tokens) with `read_api` scope
2. Create two projects:
   - `gitbackup-test-public` (public)
   - `gitbackup-test-private` (private)
3. Fork any public project and rename it to `gitbackup-test-ignore-fork` (for testing `-ignore-fork`)

### Bitbucket

1. Create an [API token](https://bitbucket.org/account/settings/app-passwords/) with `read:user:bitbucket`, `read:workspace:bitbucket`, and `read:repository:bitbucket` scopes
2. Create a workspace and two repositories:
   - `gitbackup-test-public` (public)
   - `gitbackup-test-private` (private)
3. Fork any public repo into your workspace and rename it to `gitbackup-test-ignore-fork` (for testing `-ignore-fork`)

### Forgejo (Codeberg)

1. Create an [access token](https://codeberg.org/user/settings/applications) with `read:repository` permission
2. Create two repositories:
   - `gitbackup-test-public` (public)
   - `gitbackup-test-private` (private)
3. Fork any public repo and rename it to `gitbackup-test-ignore-fork` (for testing `-ignore-fork`)

## Environment Setup

```
cp test/.env.example test/.env
```

Fill in your tokens in `test/.env`.

## Running the Tests

From the repository root:

```
bash test/integration-test.sh
```

Any services with missing environment variables will be skipped automatically.
