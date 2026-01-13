# Repoman

Repoman is a CLI tool designed for instructors to manage student repositories. It interacts with a web application to clone, sync, and track repositories for specific courses and assignments.

> [!IMPORTANT]
> **Disclosure and warning:** This program is 100% vibe-coded. It may be full of subtle bugs and jank.

## Features

- **Workspace-Based Workflow**: Link any directory to a specific Course/Assignment using `repoman init`.
- **Guided Setup**: Exploratory initialization with interactive menus to select courses and assignments.
- **Combined Dashboard**: A unified `status` view showing server-side repo lists alongside local Git status (branch, modified files, ahead/behind).
- **Secure Configuration**: Stores API keys using the system keyring (via `go-keyring`) with a secure file fallback.
- **Concurrent Syncing**: Concurrently clones or pulls multiple repositories.
- **Self-Update**: Automatically checks for and installs the latest version from GitHub Releases.

## Installation

### Pre-built Binaries

You can download the latest pre-built binary for your operating system from the [GitHub Releases](https://github.com/liffiton/repoman/releases) page. Download the file for your platform (e.g., `repoman-linux-amd64` or `repoman-windows-amd64.exe`), rename it to `repoman` (or `repoman.exe`), and place it in your system's PATH.

On Mac and Linux, you can automate this using the install script:
```bash
curl -sSL https://raw.githubusercontent.com/liffiton/repoman/main/install.sh | bash
```

### From Source

Ensure you have Go installed (version 1.24 or later).

```bash
git clone https://github.com/liffiton/repoman.git
cd repoman
go install .
```

## Usage

### 1. Authenticate
Set your API key and server URL:

```bash
repoman auth
```

### 2. Initialize a Workspace
Go to the directory where you want to store student repositories and run:

```bash
repoman init
```
*Follow the interactive prompts to select your Course and Assignment.*

### 3. Sync Repositories
Clone or update all student repositories for the current assignment:

```bash
repoman sync
```

### 4. Status Dashboard
View the status of all student repositories (both local and server-side):

```bash
repoman status
```

### 5. Self-Update
Update the `repoman` binary to the latest version:

```bash
repoman update
```

## Development & Contributing

If you want to build Repoman from source, run tests, or contribute to the project, please see the [Development Guide](DEVELOPMENT.md).

## License
MIT
