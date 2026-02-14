// Package git provides wrappers for git operations.
package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// runGitCmd executes a git command with the given arguments.
// It enforces non-interactive behavior and strict host key checking.
// The acceptNewHosts flag controls whether new host keys are accepted automatically.
func runGitCmd(ctx context.Context, acceptNewHosts bool, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "git", args...)

	strictHostKeyChecking := "yes"
	if acceptNewHosts {
		strictHostKeyChecking = "accept-new"
	}

	sshOptions := fmt.Sprintf("-o StrictHostKeyChecking=%s -o BatchMode=yes -o ConnectTimeout=10", strictHostKeyChecking)

	var sshCommand string
	if existingSSH := os.Getenv("GIT_SSH_COMMAND"); existingSSH != "" {
		sshCommand = existingSSH + " " + sshOptions
	} else {
		sshCommand = "ssh " + sshOptions
	}

	cmd.Env = append(os.Environ(),
		"GIT_TERMINAL_PROMPT=0",
		fmt.Sprintf("GIT_SSH_COMMAND=%s", sshCommand))

	return cmd.CombinedOutput()
}

const (
	defaultCloneTimeout = 5 * time.Minute
	defaultPullTimeout  = 2 * time.Minute
)

// Sync ensures the repository at the given URL is present and up-to-date at the given path.
// It uses the SSH URL by default unless useHTTP is true.
func Sync(url, path string, useHTTP bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultCloneTimeout)
	defer cancel()
	return SyncCtx(ctx, url, path, useHTTP)
}

// SyncCtx ensures the repository at the given URL is present and up-to-date at the given path.
// It uses the SSH URL by default unless useHTTP is true.
// Uses the provided context for timeout/cancellation control.
func SyncCtx(ctx context.Context, url, path string, useHTTP bool) error {
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
		return PullCtx(ctx, path)
	} else if !os.IsNotExist(err) {
		return err
	}

	return CloneCtx(ctx, url, path, useHTTP)
}

// Clone clones a repository.
// It uses the SSH URL by default unless useHTTP is true.
func Clone(url, path string, useHTTP bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultCloneTimeout)
	defer cancel()
	return CloneCtx(ctx, url, path, useHTTP)
}

// CloneCtx clones a repository.
// It uses the SSH URL by default unless useHTTP is true.
// Uses the provided context for timeout/cancellation control.
func CloneCtx(ctx context.Context, url, path string, useHTTP bool) error {
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
	output, err := runGitCmd(ctx, true, "clone", url, path)
	if err != nil {
		return wrapGitError(err, output, "git clone")
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
	ctx, cancel := context.WithTimeout(context.Background(), defaultPullTimeout)
	defer cancel()
	return PullCtx(ctx, path)
}

// PullCtx pulls changes in an existing repository.
// Uses the provided context for timeout/cancellation control.
func PullCtx(ctx context.Context, path string) error {
	output, err := runGitCmd(ctx, false, "-C", path, "pull")
	if err != nil {
		// Check if it's an empty repository
		if count, countErr := GetCommitCountCtx(ctx, path); countErr == nil && count == 0 {
			return fmt.Errorf("repository is empty")
		}
		return wrapGitError(err, output, "git pull")
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
	return GetStatusCtx(context.Background(), path)
}

// GetStatusCtx returns the current branch and a summary of the status.
// Uses the provided context for timeout/cancellation control.
func GetStatusCtx(ctx context.Context, path string) (branch, summary string, err error) {
	branch = GetBranchCtx(ctx, path)

	// Check if the repository is empty
	count, err := GetCommitCountCtx(ctx, path)
	if err != nil {
		return branch, "", fmt.Errorf("failed to get commit count: %w", err)
	}
	if count == 0 {
		return branch, "Empty repo.", nil
	}

	// Get status summary
	out, err := runGitCmd(ctx, false, "-C", path, "status", "--short")
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
	return GetCommitCountCtx(context.Background(), path)
}

// GetCommitCountCtx returns the number of commits in the repository.
// Uses the provided context for timeout/cancellation control.
func GetCommitCountCtx(ctx context.Context, path string) (int, error) {
	out, err := runGitCmd(ctx, false, "-C", path, "rev-list", "--all", "--count")
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
	return GetBranchCtx(context.Background(), path)
}

// GetBranchCtx returns the name of the current branch.
// It is more robust than 'git rev-parse --abbrev-ref HEAD' as it works on empty repositories.
// Uses the provided context for timeout/cancellation control.
func GetBranchCtx(ctx context.Context, path string) string {
	// Try symbolic-ref first (works on empty repos)
	out, err := runGitCmd(ctx, false, "-C", path, "symbolic-ref", "--short", "HEAD")
	if err == nil {
		return strings.TrimSpace(string(out))
	}

	// Fallback to rev-parse for detached HEAD
	out, err = runGitCmd(ctx, false, "-C", path, "rev-parse", "--abbrev-ref", "HEAD")
	if err == nil {
		return strings.TrimSpace(string(out))
	}

	return "Unknown"
}

// Fetch fetches from the remote.
func Fetch(path string) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultPullTimeout)
	defer cancel()
	return FetchCtx(ctx, path)
}

// FetchCtx fetches from the remote.
// Uses the provided context for timeout/cancellation control.
func FetchCtx(ctx context.Context, path string) error {
	output, err := runGitCmd(ctx, false, "-C", path, "fetch")
	if err != nil {
		return wrapGitError(err, output, "git fetch")
	}
	return nil
}

// GetSyncState returns whether the local repo is ahead, behind, or even with the remote.
func GetSyncState(path string) (string, error) {
	return GetSyncStateCtx(context.Background(), path)
}

// GetSyncStateCtx returns whether the local repo is ahead, behind, or even with the remote.
// Uses the provided context for timeout/cancellation control.
func GetSyncStateCtx(ctx context.Context, path string) (string, error) {
	// If the repository is empty, sync state doesn't really apply in the same way
	count, err := GetCommitCountCtx(ctx, path)
	if err != nil {
		return "Unknown", err
	}
	if count == 0 {
		return "-", nil
	}

	out, err := runGitCmd(ctx, false, "-C", path, "rev-list", "--left-right", "--count", "HEAD...@{u}")
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
// If the repository has no commits, it returns a zero time and no error.
func GetLastCommitTime(path string) (time.Time, error) {
	return GetLastCommitTimeCtx(context.Background(), path)
}

// GetLastCommitTimeCtx returns the time of the most recent commit in the repository (across all branches).
// If the repository has no commits, it returns a zero time and no error.
// Uses the provided context for timeout/cancellation control.
func GetLastCommitTimeCtx(ctx context.Context, path string) (time.Time, error) {
	out, err := runGitCmd(ctx, false, "-C", path, "log", "-1", "--format=%at", "--all")
	if err != nil {
		// If it's an empty repo or some other error, check if it's actually empty
		if count, countErr := GetCommitCountCtx(ctx, path); countErr == nil && count == 0 {
			return time.Time{}, nil
		}
		return time.Time{}, err
	}
	s := strings.TrimSpace(string(out))
	if s == "" {
		return time.Time{}, nil
	}
	sec, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse commit time: %w", err)
	}
	return time.Unix(sec, 0), nil
}

func wrapGitError(err error, output []byte, operation string) error {
	outputStr := string(output)
	errMsg := err.Error()

	hint := ""

	switch {
	case strings.Contains(outputStr, "Permission denied, please try again"),
		strings.Contains(outputStr, "Permission denied (publickey)"),
		strings.Contains(outputStr, "publickey"),
		strings.Contains(errMsg, "exit status 255"):
		hint = "SSH authentication failed. Ensure your SSH key is added to ssh-agent (ssh-add) and your public key is registered with the remote server."

	case strings.Contains(outputStr, "Authentication failed"),
		strings.Contains(outputStr, "401"),
		strings.Contains(outputStr, "403"),
		strings.Contains(outputStr, "Logon failed"):
		hint = "HTTP authentication failed. Configure a Git credential helper or check your credentials."

	case strings.Contains(outputStr, "Connection refused"),
		strings.Contains(outputStr, "Connection timed out"):
		hint = "Connection refused/timed out. The remote server may be down or unreachable."

	case strings.Contains(outputStr, "Host key verification failed"):
		hint = "SSH host key verification failed. This is a security issue - investigate before proceeding."

	case strings.Contains(outputStr, "fatal: bad object") || strings.Contains(outputStr, "fatal: remote error"):
		hint = "Remote error - the repository may not exist or you may not have access."
	}

	if hint != "" {
		return fmt.Errorf("%s failed: %w\n  hint: %s", operation, err, hint)
	}
	return fmt.Errorf("%s failed: %w", operation, err)
}
