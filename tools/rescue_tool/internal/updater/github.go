package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/naidrahiqa/epitaph_rescue/internal/logger"
)

const (
	GitHubAPIURL = "https://api.github.com/repos/naidrahiqa/epitaph_rescue/releases/latest"
	UserAgent    = "EpitaphRescue/1.0"
)

// ReleaseInfo holds info about a GitHub release
type ReleaseInfo struct {
	TagName     string `json:"tag_name"`
	Name        string `json:"name"`
	Body        string `json:"body"`
	HTMLURL     string `json:"html_url"`
	PublishedAt string `json:"published_at"`
	Assets      []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
		Size               int64  `json:"size"`
	} `json:"assets"`
}

// UpdateChecker checks for new releases on GitHub
type UpdateChecker struct {
	currentVersion string
	latestRelease  *ReleaseInfo
}

// NewUpdateChecker creates a new update checker
func NewUpdateChecker(currentVersion string) *UpdateChecker {
	return &UpdateChecker{
		currentVersion: currentVersion,
	}
}

// Check fetches the latest release info from GitHub
func (uc *UpdateChecker) Check() (*ReleaseInfo, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("GET", GitHubAPIURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gagal cek update: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("belum ada release yang dipublish")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var release ReleaseInfo
	if err := json.Unmarshal(body, &release); err != nil {
		return nil, fmt.Errorf("gagal parse response: %w", err)
	}

	uc.latestRelease = &release
	logger.Info("Update check: latest=%s, current=%s", release.TagName, uc.currentVersion)
	return &release, nil
}

// IsUpdateAvailable returns true if the latest release is newer than current version
func (uc *UpdateChecker) IsUpdateAvailable() bool {
	if uc.latestRelease == nil {
		return false
	}
	latestTag := strings.TrimPrefix(uc.latestRelease.TagName, "v")
	currentTag := strings.TrimPrefix(uc.currentVersion, "v")
	return latestTag != currentTag && latestTag > currentTag
}

// GetDownloadURL returns the download URL for the EpitaphRescue.exe asset
func (uc *UpdateChecker) GetDownloadURL() string {
	if uc.latestRelease == nil {
		return ""
	}
	for _, asset := range uc.latestRelease.Assets {
		if strings.Contains(strings.ToLower(asset.Name), "epitaphrescue") &&
			strings.HasSuffix(strings.ToLower(asset.Name), ".exe") {
			return asset.BrowserDownloadURL
		}
	}
	// Fallback to release page
	return uc.latestRelease.HTMLURL
}

// LatestRelease returns the cached latest release info
func (uc *UpdateChecker) LatestRelease() *ReleaseInfo {
	return uc.latestRelease
}
