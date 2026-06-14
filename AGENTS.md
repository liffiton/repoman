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
  - *Example:* `go test -v -run ^TestSync$ ./internal/git`
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

### Error Handling
- Use `errors.New` for static error messages with no formatting.
- Use `fmt.Errorf` for formatted errors, including `fmt.Errorf("context: %w", err)` to wrap errors with additional context.
- **When to wrap:** Wrap at public function boundaries to identify which operation failed (e.g., `"failed to fetch courses: %w"`). Don't wrap when passing through immediately, when the underlying message is already clear (e.g., `os.Open` includes the filename), or at every level — redundant chains like `"failed to X: failed to Y: failed to Z: %w"` add noise.

### Security
- **API Keys:** Never hardcode secrets. Use the OS-native keyring via `github.com/zalando/go-keyring` for secure storage. Fallback to a config file with `0600` permissions in the user's config directory (`os.UserConfigDir()`).
- **Git URLs:** Validate URLs before passing them to shell commands to prevent command injection.

---

## 3. Project Architecture

### Components
- `main.go`: Root entry point that calls `cmd.Execute()`.
- `cmd/`: CLI command implementations using `cobra`. Root command is in `root.go`; subcommands include `init.go`, `auth.go`, `sync.go`, `status.go`, and `update.go`. Shared utilities are in `util.go`.
- `internal/api`: Client logic for the web application interface (`client.go`).
- `internal/git`: Wrappers for git operations (`git.go`) and concurrent management (`manager.go`).
- `internal/config`: Configuration management (`config.go`) for user settings and workspace state.
- `internal/update`: Self-update logic using GitHub Releases.
- `internal/ui`: Terminal UI helpers using `pterm` for progress bars, styled text, and interactive prompts.

### Key Files & Responsibilities
- `cmd/root.go`: Root command definition. (No root-level global flags are defined; flags are scoped to individual subcommands.)
- `cmd/sync.go`: Implementation of the `sync` command, uses `internal/git` to clone/pull repos.
- `cmd/status.go`: Implementation of the `status` command, checks local repo states.
- `internal/api/client.go`: Defines the `Repo` struct (`Name`, `URL`), `Course` struct (`ID`, `Name`), `Assignment` struct (`ID`, `Name`), the `Client` for API communication, and `extractRepoName` for URL name extraction.
- `internal/git/git.go`: Contains git operation wrappers including `Sync`, `Clone`, `Pull`, `Fetch`, `GetStatus`, `GetBranch`, `GetCommitCount`, `GetSyncState`, `GetLastCommitTime`, and URL utilities `ToSSH` and `ToHTTP`. All git operations also have `*Ctx` variants for context-aware cancellation.
- `internal/git/manager.go`: Defines `RepoInfo` (`Name`, `URL`, `Path`, `UseHTTP`), `RepoStatus` (`Name`, `Status`, `Error`, `Branch`, `CommitCount`, `SyncState`, `LastCommitTime`), the `Manager` for parallel execution, and status constants: `StatusMissing`, `StatusError`, `StateUnknown`, `StateStale`, `StateSynced`.
- `internal/config/config.go`: Handles the user config file via `os.UserConfigDir()` (e.g., `~/.config/repoman/config.json` on Linux) and local `.repoman.json`.

### Self-Update Strategy
- Releases should be hosted on **GitHub Releases**.
- The tool should check for the latest tag/release and download the appropriate binary for the current OS/Arch.
- Use a library like `github.com/minio/selfupdate` to handle the binary replacement safely.

### Git Operations
- Use `git` commands via `os/exec` for maximum compatibility and performance.  If deeper inspection is needed, ask about potentially adding a library like `go-git`.
- Efficiently update repos by using goroutines for concurrent pulls/clones, but limit concurrency to avoid overwhelming the system.

## 4. Documentation
- All exported symbols must have a doc comment.
- Keep `README.md` updated with usage examples.
- Use `AGENTS.md` (this file) to record project-specific instructions for AI tools.

## 5. Tips for AI Agents

### Editing and Imports
- **Prefer `goimports`:** When you need to add or remove imports in Go files, focus on modifying the code itself, then run `goimports -w .`. This is much safer than manually editing the `import` block, which can easily lead to accidentally removing required dependencies.
- **Block Integrity:** When using the `edit` tool to modify multi-line structures (like `import`, `struct`, or `interface` blocks), include the entire block in your `oldString`. This ensures you have the correct context and prevents accidental truncation of lines in the middle.
- **Verify Symbols:** If an edit results in "undefined" symbol errors in the LSP or build output, double-check that you didn't accidentally overwrite an import or a variable definition in a neighboring line.
