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

You can download the latest pre-built binary for your operating system from the [GitHub Releases](https://github.com/liffiton/repoman/releases) page. Download the file for your platform (e.g., `repoman-linux-amd64` or `repoman-windows-amd64.exe`), rename it to `repoman` (or `repoman.exe`), and place it in your system's PATH.

On Mac and Linux (including WSL), you can **automate** this using the install script:
```bash
curl -sSL https://raw.githubusercontent.com/liffiton/repoman/main/install.sh | bash
```

## Usage

### 1. Authenticate & Configure

```bash
~ $ repoman auth
Your API key can be found in the Settings page of the Class Repo Manager web application.
Enter API Key: my-api-key
Enter Base URL (default, if nothing entered: [https://crm.unsatisfiable.net]): 

Authentication configured successfully!
API Key: Saved securely in the system keyring.
Base URL: https://crm.unsatisfiable.net (using default, no config file created)
```

### 2. Initialize a Workspace Directory

Go to the directory in which you want to clone and store student repositories.
Follow the interactive prompts to select a Course and Assignment.

```bash
~ $ cd cs101/lab1/

~/cs101/lab1 $ repoman init
? Select a course: CS101
? Select an assignment: Lab 1
Workspace initialized for CS101 - Lab 1
```

### 3. Sync Repositories
Clone or update all student repositories for the current workspace/assignment.

```bash
~/cs101/lab1 $ repoman sync
Syncing 8 repositories for Lab 1...
 100% |█████████████████████████████████████████████████████████████████| (8/8, 6 it/s)        

Sync complete. 8/8 repositories synced successfully.
```

### 4. Status Dashboard

```bash
~/cs101/lab1 $ repoman status
Status for CS101 - Lab 1

Checking status 100% |█████████████████████████████████████████████████| (8/8, 11 it/s)        

STUDENT/REPO       BRANCH   LOCAL STATUS   SYNC STATE
Amara              main     Clean          Synced
Dmitri             main     Clean          Synced
Kenji              main     Clean          Synced
Lucia              main     Clean          Synced
Priya              main     Clean          Synced
Rafael             main     Clean          Synced
Soren              main     Clean          Synced
Yasmin             main     Clean          Synced
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
