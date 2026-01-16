// Package git provides wrappers for git operations.
package git

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Run a git command with the given arguments passed to the git CLI command.
// Strictly checks SSH host keys, failing without a prompt if they are missing
// or don't match.
func runGitCommand(args ...string) ([]byte, error) {
	return runGitCommandAcceptNewHosts(false, args...)
}

// Run a git command with the given arguments passed to the git CLI command.
// Allows accepting new host keys without a prompt (if acceptNewHosts=true).
// Otherwise, strictly checks SSH host keys, failing without a prompt if they
// are missing or don't match.
func runGitCommandAcceptNewHosts(acceptNewHosts bool, args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)

	strictHostKeyChecking := "yes"
	if acceptNewHosts {
		strictHostKeyChecking = "accept-new"
	}

	cmd.Env = append(os.Environ(),
		fmt.Sprintf("GIT_SSH_COMMAND=ssh -o StrictHostKeyChecking=%s", strictHostKeyChecking))

	return cmd.CombinedOutput()
}

// Sync ensures the repository at the given URL is present and up-to-date at the given path.
// It uses the SSH URL by default unless useHTTP is true.
func Sync(url, path string, useHTTP bool) error {
	if useHTTP {
		url = ToHTTP(url)
	} else {
		url = ToSSH(url)
	}

	if info, err := os.Stat(path); err == nil {
		if !info.IsDir() {
			return fmt.Errorf("path %s exists but is not a directory", path)
		}
		// Check if it is a git repo
		if _, err := os.Stat(fmt.Sprintf("%s/.git", path)); err != nil {
			return fmt.Errorf("path %s exists but is not a git repository", path)
		}
		return Pull(path)
	} else if !os.IsNotExist(err) {
		return err
	}

	return Clone(url, path, useHTTP)
}

// Clone clones a repository.
// It uses the SSH URL by default unless useHTTP is true.
func Clone(url, path string, useHTTP bool) error {
	if useHTTP {
		url = ToHTTP(url)
	} else {
		url = ToSSH(url)
	}

	if err := validateURL(url); err != nil {
		return err
	}

	// Accept a new host key (only here on clone) to streamline if using this tool
	// is the first time the user has connected to the Git/SSH host.
	output, err := runGitCommandAcceptNewHosts(true, "clone", url, path)
	if err != nil {
		return fmt.Errorf("git clone failed: %w (output: %s)", err, strings.TrimSpace(string(output)))
	}
	return nil
}

// ToSSH converts an HTTP/HTTPS git URL to an SSH git URL.
// If the URL is already an SSH URL or not an HTTP URL, it is returned unchanged.
func ToSSH(url string) string {
	if strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "http://") {
		// Remove protocol
		u := url
		if strings.HasPrefix(u, "https://") {
			u = u[8:]
		} else {
			u = u[7:]
		}

		u = strings.TrimSuffix(u, "/")

		// Split host and path
		parts := strings.SplitN(u, "/", 2)
		if len(parts) == 2 {
			host := parts[0]
			repoPath := parts[1]
			if !strings.HasSuffix(repoPath, ".git") {
				repoPath += ".git"
			}
			return fmt.Sprintf("git@%s:%s", host, repoPath)
		}
	}
	return url
}

// ToHTTP converts an SSH git URL to an HTTPS git URL.
// If the URL is already an HTTPS URL or not an SSH URL, it is returned unchanged.
func ToHTTP(url string) string {
	if strings.HasPrefix(url, "git@") {
		u := strings.TrimSuffix(url[4:], ".git")
		u = strings.TrimSuffix(u, "/")
		parts := strings.SplitN(u, ":", 2)
		if len(parts) == 2 {
			host := parts[0]
			repoPath := parts[1]
			return fmt.Sprintf("https://%s/%s", host, repoPath)
		}
	} else if strings.HasPrefix(url, "ssh://git@") {
		u := strings.TrimSuffix(url[10:], ".git")
		u = strings.TrimSuffix(u, "/")
		return "https://" + u
	}
	return url
}

// ExtractRepoName extracts the repository name from a git URL.
func ExtractRepoName(url string) string {
	u := strings.TrimSuffix(url, "/")
	u = strings.TrimSuffix(u, ".git")

	// Split by '/' and take the last part
	parts := strings.Split(u, "/")
	if len(parts) > 0 {
		last := parts[len(parts)-1]
		// Handle git@github.com:repo.git (no slash between host and repo)
		if strings.Contains(last, ":") {
			subParts := strings.Split(last, ":")
			return subParts[len(subParts)-1]
		}
		return last
	}
	return ""
}

// Pull pulls changes in an existing repository.
func Pull(path string) error {
	output, err := runGitCommand("-C", path, "pull")
	if err != nil {
		// Check if it's an empty repository
		if count, countErr := GetCommitCount(path); countErr == nil && count == 0 {
			return fmt.Errorf("repository is empty")
		}
		return fmt.Errorf("git pull failed: %w (output: %s)", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func validateURL(url string) error {
	// Basic validation to prevent command injection
	if strings.Contains(url, " ") || strings.HasPrefix(url, "-") {
		return fmt.Errorf("invalid git URL: %s", url)
	}
	// We could add more robust validation here if needed
	return nil
}

// GetStatus returns the current branch and a summary of the status.
func GetStatus(path string) (branch, summary string, err error) {
	branch = GetBranch(path)

	// Check if the repository is empty
	count, err := GetCommitCount(path)
	if err != nil {
		return branch, "", fmt.Errorf("failed to get commit count: %w", err)
	}
	if count == 0 {
		return branch, "Empty repo.", nil
	}

	// Get status summary
	out, err := runGitCommand("-C", path, "status", "--short")
	if err != nil {
		return branch, "", fmt.Errorf("failed to get status: %w", err)
	}

	if len(out) == 0 {
		summary = "Clean"
	} else {
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		summary = fmt.Sprintf("%d files modified", len(lines))
	}

	return branch, summary, nil
}

// GetCommitCount returns the number of commits in the repository.
func GetCommitCount(path string) (int, error) {
	out, err := runGitCommand("-C", path, "rev-list", "--all", "--count")
	if err != nil {
		return 0, err
	}
	var count int
	_, err = fmt.Sscanf(strings.TrimSpace(string(out)), "%d", &count)
	if err != nil {
		return 0, fmt.Errorf("failed to parse commit count: %w", err)
	}
	return count, nil
}

// GetBranch returns the name of the current branch.
// It is more robust than 'git rev-parse --abbrev-ref HEAD' as it works on empty repositories.
func GetBranch(path string) string {
	// Try symbolic-ref first (works on empty repos)
	out, err := runGitCommand("-C", path, "symbolic-ref", "--short", "HEAD")
	if err == nil {
		return strings.TrimSpace(string(out))
	}

	// Fallback to rev-parse for detached HEAD
	out, err = runGitCommand("-C", path, "rev-parse", "--abbrev-ref", "HEAD")
	if err == nil {
		return strings.TrimSpace(string(out))
	}

	return "Unknown"
}

// Fetch fetches from the remote.
func Fetch(path string) error {
	_, err := runGitCommand("-C", path, "fetch")
	if err != nil {
		return fmt.Errorf("git fetch failed: %w", err)
	}
	return nil
}

// GetSyncState returns whether the local repo is ahead, behind, or even with the remote.
func GetSyncState(path string) (string, error) {
	// If the repository is empty, sync state doesn't really apply in the same way
	count, err := GetCommitCount(path)
	if err != nil {
		return "Unknown", err
	}
	if count == 0 {
		return "-", nil
	}

	out, err := runGitCommand("-C", path, "rev-list", "--left-right", "--count", "HEAD...@{u}")
	if err != nil {
		return "Unknown", fmt.Errorf("failed to get sync state: %w", err)
	}

	parts := strings.Fields(string(out))
	if len(parts) != 2 {
		return "Unknown", fmt.Errorf("unexpected output from rev-list: %s", string(out))
	}

	ahead := parts[0]
	behind := parts[1]

	if ahead == "0" && behind == "0" {
		return "Synced", nil
	}
	if ahead != "0" && behind != "0" {
		return fmt.Sprintf("Diverged (+%s, -%s)", ahead, behind), nil
	}
	if ahead != "0" {
		return fmt.Sprintf("Ahead (+%s)", ahead), nil
	}
	return fmt.Sprintf("Behind (-%s)", behind), nil
}

// GetLastCommitTime returns the time of the most recent commit in the repository (across all branches).
func GetLastCommitTime(path string) (time.Time, error) {
	out, err := runGitCommand("-C", path, "log", "-1", "--format=%at", "--all")
	if err != nil {
		return time.Time{}, err
	}
	s := strings.TrimSpace(string(out))
	if s == "" {
		return time.Time{}, fmt.Errorf("no commits found")
	}
	sec, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse commit time: %w", err)
	}
	return time.Unix(sec, 0), nil
}
