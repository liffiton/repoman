package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/zalando/go-keyring"
)

func TestMain(m *testing.M) {
	keyring.MockInit()
	os.Exit(m.Run())
}

func TestConfigLoadSave(t *testing.T) {
	// Override UserConfigDir for testing
	tmpDir, err := os.MkdirTemp("", "repoman-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// We can't easily override os.UserConfigDir() globally without changing the environment variable
	originalConfigDir := os.Getenv("XDG_CONFIG_HOME")
	defer func() { _ = os.Setenv("XDG_CONFIG_HOME", originalConfigDir) }()
	_ = os.Setenv("XDG_CONFIG_HOME", tmpDir)

	// Also for macOS/Windows, but we are on Linux based on env
	_ = os.Setenv("HOME", tmpDir)

	cfg := &Config{
		APIKey: "test-api-key",
	}

	_, err = cfg.Save()
	if err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	loadedCfg, err := Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if loadedCfg.APIKey != "test-api-key" {
		t.Errorf("expected APIKey 'test-api-key', got '%s'", loadedCfg.APIKey)
	}
}

func TestEnsureConfigDir(t *testing.T) {
	dir, err := EnsureConfigDir()
	if err != nil {
		t.Fatalf("EnsureConfigDir failed: %v", err)
	}
	if dir == "" {
		t.Error("EnsureConfigDir returned empty string")
	}

	// Verify it's a directory
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}
	if !info.IsDir() {
		t.Error("EnsureConfigDir path is not a directory")
	}
}

func TestFindWorkspaceRoot(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "repoman-workspace-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a workspace file in the temp dir
	wsFile := filepath.Join(tmpDir, workspaceFileName)
	if err := os.WriteFile(wsFile, []byte("{}"), 0o600); err != nil {
		t.Fatalf("failed to create workspace file: %v", err)
	}

	// Create a subdirectory
	subDir := filepath.Join(tmpDir, "sub", "dir")
	if err := os.MkdirAll(subDir, 0o700); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	// Change to the subdirectory
	oldWd, _ := os.Getwd()
	if err := os.Chdir(subDir); err != nil {
		t.Fatalf("failed to change to subdir: %v", err)
	}
	defer func() { _ = os.Chdir(oldWd) }()

	// Find the workspace root
	root, err := FindWorkspaceRoot()
	if err != nil {
		t.Fatalf("FindWorkspaceRoot failed: %v", err)
	}

	absTmpDir, _ := filepath.Abs(tmpDir)
	absRoot, _ := filepath.Abs(root)
	if absRoot != absTmpDir {
		t.Errorf("expected root %s, got %s", absTmpDir, absRoot)
	}
}
