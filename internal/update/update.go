// Package update provides self-update logic for the repoman tool.
package update

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"

	"github.com/minio/selfupdate"
	"github.com/schollz/progressbar/v3"
)

const (
	githubOwner = "liffiton"
	githubRepo  = "repoman"
)

// Release represents a GitHub release.
type Release struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

// Asset represents a GitHub release asset.
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// CheckAndUpdate checks for a new version on GitHub and performs the update if available.
func CheckAndUpdate(currentVersion string) (bool, error) {
	// #nosec G107
	resp, err := http.Get(fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", githubOwner, githubRepo))
	if err != nil {
		return false, fmt.Errorf("failed to check for updates: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return false, nil // No releases yet
	}
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("unexpected status code checking for updates: %d", resp.StatusCode)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return false, fmt.Errorf("failed to decode release info: %w", err)
	}

	if release.TagName == currentVersion {
		return false, nil // Up to date
	}

	// Find the asset for the current OS and Arch
	// Expecting naming like repoman-linux-amd64 or repoman-windows-amd64.exe
	extension := ""
	if runtime.GOOS == "windows" {
		extension = ".exe"
	}
	targetAsset := fmt.Sprintf("repoman-%s-%s%s", runtime.GOOS, runtime.GOARCH, extension)
	var downloadURL string
	for _, asset := range release.Assets {
		if asset.Name == targetAsset {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		return false, fmt.Errorf("no suitable asset found in latest release for %s", targetAsset)
	}

	if err := doUpdate(downloadURL); err != nil {
		return false, fmt.Errorf("failed to apply update: %w", err)
	}

	return true, nil
}

func doUpdate(url string) error {
	// #nosec G107
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code downloading update: %d", resp.StatusCode)
	}

	bar := progressbar.DefaultBytes(
		resp.ContentLength,
		"Downloading update",
	)

	return selfupdate.Apply(io.TeeReader(resp.Body, bar), selfupdate.Options{})
}
