# Development Guide

This document provides instructions for developers who want to contribute to Repoman or build it from source.

## Prerequisites

- **Go**: Version 1.24 or later.
- **Git**: Required for cloning the repository and for Repoman's core functionality.
- **golangci-lint**: Recommended for local linting.

## Getting Started

1.  **Clone the repository**:
    ```bash
    git clone https://github.com/liffiton/repoman.git
    cd repoman
    ```

2.  **Install dependencies**:
    ```bash
    go mod download
    ```

## Build and Install

### Local Build
To build the binary in the current directory:
```bash
go build -o repoman .
```

### Install to GOBIN
To install the binary to your `$GOPATH/bin` (or `$HOME/go/bin`):
```bash
go install .
```

## Testing

Run all tests with:
```bash
go test ./...
```

Run tests with coverage:
```bash
go test -cover ./...
```

## Linting

We use `golangci-lint` for code quality. To install it locally:
```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

Run the linter:
```bash
golangci-lint run
```

## Project Structure

- `main.go`: Root entry point.
- `cmd/`: CLI command package (using Cobra).
- `internal/api/`: Client for the Repoman web application.
- `internal/config/`: Configuration management (user settings and workspace state).
- `internal/git/`: Git operation wrappers and concurrent management.
- `internal/update/`: Self-update logic.

## Release Process

Repoman uses **GoReleaser** and **GitHub Actions** for automated releases.

### How it works
1.  **CI**: Every push or Pull Request to `main` triggers a CI workflow (`.github/workflows/ci.yml`) that runs linters and tests.
2.  **Release**: Pushing a git tag starting with `v` (e.g., `v1.0.0`) triggers the release workflow (`.github/workflows/release.yml`).
3.  **GoReleaser**: The release workflow uses GoReleaser to:
    - Build cross-platform binaries (Linux, macOS, Windows).
    - Inject the version number into the binary.
    - Generate a changelog.
    - Create a GitHub Release and upload the assets.

### To Create a New Release
1.  Ensure you are on the `main` branch and it is up to date.
2.  Tag the commit:
    ```bash
    git tag -a v1.0.0 -m "Release v1.0.0"
    ```
3.  Push the tag:
    ```bash
    git push origin v1.0.0
    ```
4.  Monitor the "Actions" tab on GitHub. Once finished, the new version will be available on the "Releases" page, and the `repoman update` command will be able to detect it.
