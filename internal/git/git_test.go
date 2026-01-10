package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestSync(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "repoman-git-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a source repo
	srcRepo := filepath.Join(tmpDir, "src")
	if err := os.MkdirAll(srcRepo, 0o750); err != nil {
		t.Fatalf("failed to create src repo dir: %v", err)
	}

	runGit := func(dir string, args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git command failed: %v (output: %s)", err, string(output))
		}
	}

	runGit(srcRepo, "init", "-b", "main")
	runGit(srcRepo, "config", "user.email", "test@example.com")
	runGit(srcRepo, "config", "user.name", "Test User")
	if err := os.WriteFile(filepath.Join(srcRepo, "test.txt"), []byte("hello"), 0o600); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	runGit(srcRepo, "add", "test.txt")
	runGit(srcRepo, "commit", "-m", "initial commit")

	destRepo := filepath.Join(tmpDir, "dest")

	// First Sync (Clone)
	if err := Sync(srcRepo, destRepo, false); err != nil {
		t.Fatalf("First Sync (Clone) failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(destRepo, "test.txt")); err != nil {
		t.Errorf("cloned repo missing test.txt: %v", err)
	}

	// Second Sync (Pull)
	if err := os.WriteFile(filepath.Join(srcRepo, "test2.txt"), []byte("world"), 0o600); err != nil {
		t.Fatalf("failed to write second file: %v", err)
	}
	runGit(srcRepo, "add", "test2.txt")
	runGit(srcRepo, "commit", "-m", "second commit")

	if err := Sync(srcRepo, destRepo, false); err != nil {
		t.Fatalf("Second Sync (Pull) failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(destRepo, "test2.txt")); err != nil {
		t.Errorf("pulled repo missing test2.txt: %v", err)
	}
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		url     string
		wantErr bool
	}{
		{"https://github.com/user/repo", false},
		{"git@github.com:user/repo.git", false},
		{"https://github.com/user/repo --option", true},
		{"-o github.com/user/repo", true},
		{"ssh://git@github.com/user/repo", false},
	}

	for _, tt := range tests {
		err := validateURL(tt.url)
		if (err != nil) != tt.wantErr {
			t.Errorf("validateURL(%q) error = %v, wantErr %v", tt.url, err, tt.wantErr)
		}
	}
}

func TestURLConversion(t *testing.T) {
	tests := []struct {
		url      string
		wantSSH  string
		wantHTTP string
	}{
		{
			url:      "https://github.com/user/repo",
			wantSSH:  "git@github.com:user/repo.git",
			wantHTTP: "https://github.com/user/repo",
		},
		{
			url:      "https://github.com/user/repo.git",
			wantSSH:  "git@github.com:user/repo.git",
			wantHTTP: "https://github.com/user/repo.git",
		},
		{
			url:      "git@github.com:user/repo.git",
			wantSSH:  "git@github.com:user/repo.git",
			wantHTTP: "https://github.com/user/repo",
		},
		{
			url:      "ssh://git@github.com/user/repo.git",
			wantSSH:  "ssh://git@github.com/user/repo.git",
			wantHTTP: "https://github.com/user/repo",
		},
	}

	for _, tt := range tests {
		if got := ToSSH(tt.url); got != tt.wantSSH {
			t.Errorf("ToSSH(%q) = %q, want %q", tt.url, got, tt.wantSSH)
		}
		if got := ToHTTP(tt.url); got != tt.wantHTTP {
			t.Errorf("ToHTTP(%q) = %q, want %q", tt.url, got, tt.wantHTTP)
		}
	}
}

func TestExtractRepoName(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{"https://github.com/user/repo", "repo"},
		{"https://github.com/user/repo.git", "repo"},
		{"git@github.com:user/repo.git", "repo"},
		{"git@github.com:repo.git", "repo"},
		{"ssh://git@github.com/user/repo.git", "repo"},
		{"https://github.com/user/repo/", "repo"},
	}

	for _, tt := range tests {
		if got := ExtractRepoName(tt.url); got != tt.want {
			t.Errorf("ExtractRepoName(%q) = %q, want %q", tt.url, got, tt.want)
		}
	}
}
