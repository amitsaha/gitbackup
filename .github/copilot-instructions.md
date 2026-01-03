# Copilot Instructions for gitbackup

## Project Overview

`gitbackup` is a command-line tool written in Go that backs up Git repositories from GitHub, GitLab, and Bitbucket. It supports two operation modes:
1. Creating clones of Git repositories (all services)
2. Creating user migrations using GitHub's Migration API (GitHub only)

## Architecture

- **Main entry point**: `main.go` - Initializes configuration, validates it, creates clients, and dispatches to appropriate handlers
- **Configuration**: `config.go` and `options.go` - Defines command-line flags and application configuration
- **Client creation**: `client.go` - Creates service-specific API clients (GitHub, GitLab, Bitbucket)
- **Repository handling**: `repositories.go` - Fetches repository lists from different services
- **Backup operations**: `backup.go` - Handles git clone and update operations
- **GitHub-specific features**:
  - `github.go` - GitHub API interactions
  - `github_create_user_migration.go` - Creates and downloads user migrations
  - `github_list_user_migrations.go` - Lists available migrations
  - `user_data.go` - Migration-related utilities

## Coding Standards

### Go Conventions
- Follow standard Go formatting (use `gofmt`)
- Use Go 1.25 features and syntax
- Keep functions focused and single-purpose
- Use descriptive variable names (e.g., `backupDir`, `ignorePrivate`)

### Error Handling
- Return errors, don't panic (except for fatal initialization errors)
- Use `log.Fatal()` or `log.Fatalf()` for unrecoverable errors in main execution path
- Provide context in error messages

### Testing
- Write table-driven tests where appropriate
- Use test helpers for setup/teardown (see `repositories_test.go`)
- Mock external dependencies (HTTP servers for API testing)
- Store expected test outputs in `testdata/` directory with `.golden` files
- Run tests with `go test` (no additional flags needed)

### Code Style
- Use global variables sparingly (mainly for dependency injection in tests)
- Inject dependencies for testability (e.g., `execCommand`, `appFS`)
- Use meaningful struct names with context (e.g., `appConfig`, not just `config`)
- Document exported functions and types
- Use constants for magic numbers and strings

## Common Patterns

### Service Detection
```go
var knownServices = map[string]string{
    "github":    "github.com",
    "gitlab":    "gitlab.com",
    "bitbucket": "bitbucket.org",
}
```

### Command-Line Flags
- Generic flags: `-service`, `-backupdir`, `-bare`, `-ignore-private`, `-ignore-fork`
- Service-specific flags use prefixes: `-github.*`, `-gitlab.*`
- Boolean flags default to `false` unless specified

### Concurrency
- Use `sync.WaitGroup` for concurrent git operations
- Limit concurrent clones with `MaxConcurrentClones` constant
- Always defer `wg.Done()` at function start

### File Operations
- Use `afero.Fs` interface for file system operations (enables testing)
- Use `path.Join()` for path construction
- Check if directories/files exist before operations

## Testing Requirements

### Running Tests
```bash
go test                    # Run all tests
go test -v                 # Run with verbose output
```

### Test Structure
- Unit tests in `*_test.go` files alongside source
- Use `httptest.NewServer()` for mocking API endpoints
- Set environment variables for tokens in test setup
- Clean up resources in teardown functions

### Golden Files
- CLI output tests use golden files in `testdata/`
- Platform-specific golden files: `TestName.golden.windows` for Windows
- Update golden files when intentionally changing output

## Building

```bash
go build                   # Build binary
```

The resulting binary will be named `gitbackup` (or `gitbackup.exe` on Windows).

## Environment Variables

- `GITHUB_TOKEN` - GitHub personal access token
- `GITLAB_TOKEN` - GitLab personal access token  
- `BITBUCKET_USERNAME` - Bitbucket username
- `BITBUCKET_PASSWORD` - Bitbucket app password

## Git Operations

- Clone operations use `git clone` command
- Bare clones use `git clone --mirror`
- Updates use `git pull` (normal) or `git remote update --prune` (bare)
- Support both SSH and HTTPS cloning (via `-use-https-clone` flag)

## Important Notes

- The tool is designed for personal backup, not as a backup service
- Repository namespaces (user/org) are preserved in backup directory structure
- Private repositories can be excluded with `-ignore-private`
- Forked repositories can be excluded with `-ignore-fork`
- GitHub migrations download `.tar.gz` archives containing full repository data

## When Making Changes

1. **Maintain backward compatibility** - Don't break existing command-line flags
2. **Update tests** - Add or modify tests for any new functionality
3. **Update README** - Document new flags or features
4. **Test on multiple platforms** - CI runs on Linux, macOS, and Windows
5. **Follow existing patterns** - Match the style and structure of similar code
6. **Update golden files** - Regenerate if CLI output changes intentionally
