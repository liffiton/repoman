package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestSyncAll(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "repoman-manager-test-*")
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

	manager := NewManager(2)
	repos := []RepoInfo{
		{Name: "dest1", URL: srcRepo, Path: filepath.Join(tmpDir, "dest1")},
		{Name: "dest2", URL: srcRepo, Path: filepath.Join(tmpDir, "dest2")},
	}

	progressCount := 0
	errs := manager.SyncAll(repos, func() {
		progressCount++
	})

	for i, err := range errs {
		if err != nil {
			t.Errorf("repo %d failed to sync: %v", i, err)
		}
	}

	if progressCount != len(repos) {
		t.Errorf("expected progress count %d, got %d", len(repos), progressCount)
	}

	if _, err := os.Stat(filepath.Join(tmpDir, "dest1", "test.txt")); err != nil {
		t.Errorf("dest1 missing test.txt")
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "dest2", "test.txt")); err != nil {
		t.Errorf("dest2 missing test.txt")
	}
}

func TestStatusAll(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "repoman-manager-status-test-*")
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

	manager := NewManager(2)
	dest1 := filepath.Join(tmpDir, "dest1")
	// Clone dest1
	runGit(tmpDir, "clone", srcRepo, "dest1")

	repos := []RepoInfo{
		{Name: "dest1", URL: srcRepo, Path: dest1},
		{Name: "missing", URL: srcRepo, Path: filepath.Join(tmpDir, "missing")},
	}

	progressCount := 0
	statuses := manager.StatusAll(repos, false, func() {
		progressCount++
	})

	if progressCount != len(repos) {
		t.Errorf("expected progress count %d, got %d", len(repos), progressCount)
	}

	if len(statuses) != 2 {
		t.Fatalf("expected 2 statuses, got %d", len(statuses))
	}

	// Check dest1
	if statuses[0].Name != "dest1" {
		t.Errorf("expected name dest1, got %s", statuses[0].Name)
	}
	if statuses[0].Status != "Clean" {
		t.Errorf("expected status Clean, got %s", statuses[0].Status)
	}
	if statuses[0].Branch != "main" {
		t.Errorf("expected branch main, got %s", statuses[0].Branch)
	}

	// Check missing
	if statuses[1].Name != "missing" {
		t.Errorf("expected name missing, got %s", statuses[1].Name)
	}
	if statuses[1].Status != "Missing" {
		t.Errorf("expected status Missing, got %s", statuses[1].Status)
	}
}
