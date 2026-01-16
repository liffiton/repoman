# Repoman Agent Guidelines

This document provides essential information for AI agents and developers working on the Repoman project. Repoman is a Go-based CLI tool designed to manage Git repositories by interfacing with a web application.

## 1. Format, Lint, Test, and Build Commands

Repoman uses standard Go toolchain commands for quality assurance.

### Format
- `gofmt -w .` (standard Go formatting).
- `goimports -w .` to manage imports and ensure they are grouped.

### Lint
- `go mod tidy`
- `go vet ./...`
- `golangci-lint run`

### Test
- **Run all tests:** `go test ./...`
- **Run tests with coverage:** `go test -cover ./...`
- **Run a single test:** `go test -v -run ^TestName$ ./path/to/package`
  - *Example:* `go test -v -run ^TestCloneRepo$ ./internal/git`
- **Run tests in a specific file:** `go test -v ./path/to/package/file_test.go`

### Build
- **Build the binary:** `go build`

### Verification after Edits
**Mandatory:** After any code changes, agents MUST:
  1. run all formatters,
  2. run all linters,
  3. run all tests, and
  4. run `go build` to ensure no regressions or build failures were introduced and to ensure the development/testing binary is up to date.

---

## 2. Code Style Guidelines

### Naming Conventions
- **Packages:** Short, lowercase, single-word names (e.g., `git`, `config`, `api`).
- **Functions/Variables:** 
  - `camelCase` for unexported (private) symbols.
  - `PascalCase` for exported (public) symbols.
- **Interfaces:** Usually end in `-er` if they define a single method (e.g., `Cloner`, `Updater`).
- **Initialisms:** Keep them consistent (e.g., `APIKey`, not `ApiKey`; `URL`, not `Url`).

### Types and Structs
- Use structs to group related data.
- Prefer functional options for complex constructors.
- Keep struct sizes manageable and prefer composition over deep embedding.

### Error Handling
- Errors should be handled explicitly. Do not ignore errors unless there is a very good reason (and document it).
- Wrap errors with context using `fmt.Errorf("context: %w", err)`.
- Use specific error types or `errors.Is` / `errors.As` for checking sentinel errors.
- **Avoid panics** in library code; reserve them for truly unrecoverable setup issues in `main`.

### Security
- **API Keys:** Never hardcode secrets. Use the OS-native keyring via `github.com/zalando/go-keyring` for secure storage. Fallback to a config file with `0600` permissions in the user's config directory (`os.UserConfigDir()`).
- **Git URLs:** Validate URLs before passing them to shell commands to prevent command injection.

---

## 3. Project Architecture

### Components
- `main.go`: Root entry point that calls `cmd.Execute()`.
- `cmd/`: CLI command implementations using `cobra`. Root command is in `root.go`, and subcommands have their own files.
- `internal/api`: Client logic for the web application interface (`client.go`).
- `internal/git`: Wrappers for git operations (`git.go`) and concurrent management (`manager.go`).
- `internal/config`: Configuration management (`config.go`) for user settings and workspace state.
- `internal/update`: Self-update logic using GitHub Releases.

### Key Files & Responsibilities
- `main.go`: Entry point for the application.
- `cmd/root.go`: Root command and global flag definitions.
- `cmd/sync.go`: Implementation of the `sync` command, uses `internal/git` to clone/pull repos.
- `cmd/status.go`: Implementation of the `status` command, checks local repo states.
- `internal/api/client.go`: Defines the `Repo` struct which contains the default `URL` from the API.
- `internal/git/git.go`: contains `Sync`, `Clone`, and `Pull` functions.
- `internal/git/manager.go`: defines `RepoInfo` and the `Manager` for parallel execution.
- `internal/config/config.go`: Handles `~/.config/repoman/config.json` and local `.repoman.json`.

### Self-Update Strategy
- Releases should be hosted on **GitHub Releases**.
- The tool should check for the latest tag/release and download the appropriate binary for the current OS/Arch.
- Use a library like `github.com/minio/selfupdate` to handle the binary replacement safely.

### Git Operations
- Use `git` commands via `os/exec` for maximum compatibility and performance, or a library like `go-git` if deep inspection is needed.
- Efficiently update repos by using goroutines for concurrent pulls/clones, but limit concurrency to avoid overwhelming the system.

## 4. Documentation
- All exported symbols must have a doc comment.
- Keep `README.md` updated with usage examples.
- Use `AGENTS.md` (this file) to record project-specific instructions for AI tools.
