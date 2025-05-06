package updater

import (
	"fmt"
	"github.com/Masterminds/semver/v3"
	"github.com/devplaninc/devplan-cli/internal/pb/cli"
	"github.com/devplaninc/devplan-cli/internal/version"
	"google.golang.org/protobuf/encoding/protojson"
	"io"
	"net/http"
	"os"
	"runtime"
)

const (
	// baseURL is the base URL for the DigitalOcean Space where releases are stored
	baseURL = "https://devplan-cli.sfo3.digitaloceanspaces.com/releases"

	// VersionFile is the name of the file that contains the current production version
	VersionFile = "version.json"
)

// Client holds the configuration for the updater
type Client struct {
}

// GetBaseURL returns the base URL for the DigitalOcean Space
func (c *Client) GetBaseURL() string {
	return baseURL
}

// GetVersionURL returns the URL for the version file
func (c *Client) GetVersionURL() string {
	return fmt.Sprintf("%s/%s", c.GetBaseURL(), VersionFile)
}

// GetReleaseURL returns the URL for a specific release
func (c *Client) GetReleaseURL(version string) string {
	return fmt.Sprintf("%s/versions/%s", c.GetBaseURL(), version)
}

// GetBinaryURL returns the URL for the binary for the current platform
func (c *Client) GetBinaryURL(version string) string {
	binaryName := getBinaryNameForCurrentPlatform()
	return fmt.Sprintf("%s/%s", c.GetReleaseURL(version), binaryName)
}

// GetVersionConfig returns the current production version
func (c *Client) GetVersionConfig() (*cli.Version, error) {
	resp, err := http.Get(c.GetVersionURL())
	if err != nil {
		return nil, fmt.Errorf("failed to get production version: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get production version: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read production version: %w", err)
	}
	ver := &cli.Version{}
	return ver, protojson.Unmarshal(body, ver)
}

// CheckForUpdate checks if there is a newer production version available
func (c *Client) CheckForUpdate() (bool, string, error) {
	ver, err := c.GetVersionConfig()
	if err != nil {
		return false, "", err
	}

	currentVersion := version.GetVersion()
	prodVer := ver.GetProductionVersion()
	if currentVersion == "dev" {
		return true, prodVer, nil
	}

	// Compare versions using semver rules
	hasUpdate, err := isNewer(prodVer, currentVersion)
	if err != nil {
		return false, "", fmt.Errorf("failed to compare versions: %w", err)
	}

	return hasUpdate, prodVer, nil
}

// isNewer compares two semantic versions and returns true if v1 is newer than v2
func isNewer(v1, v2 string) (bool, error) {
	v1Sem, err := semver.NewVersion(v1)
	if err != nil {
		return false, fmt.Errorf("invalid version %s: %w", v1, err)
	}
	v2Sem, err := semver.NewVersion(v2)
	if err != nil {
		return false, fmt.Errorf("invalid version %s: %w", v2, err)
	}
	return v1Sem.GreaterThan(v2Sem), nil
}

// Update updates the binary to the specified version
func (c *Client) Update(targetVersion string) error {
	// Get the URL for the binary
	binaryURL := c.GetBinaryURL(targetVersion)

	// Download the binary
	resp, err := http.Get(binaryURL)
	if err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download update: %s", resp.Status)
	}

	// Create a temporary file
	tempFile, err := os.CreateTemp("", "devplan-update-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tempFile.Name())

	// Copy the downloaded binary to the temporary file
	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write update to temporary file: %w", err)
	}

	// Close the temporary file
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("failed to close temporary file: %w", err)
	}

	// Make the temporary file executable
	if err := os.Chmod(tempFile.Name(), 0755); err != nil {
		return fmt.Errorf("failed to make temporary file executable: %w", err)
	}

	// Get the path to the current executable
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get path to current executable: %w", err)
	}

	// Replace the current executable with the new one
	if runtime.GOOS == "windows" {
		// On Windows, we can't replace the running executable directly
		// So we rename the current executable and copy the new one
		bakPath := execPath + ".bak"
		if err := os.Rename(execPath, bakPath); err != nil {
			return fmt.Errorf("failed to rename current executable: %w", err)
		}

		if err := copyFile(tempFile.Name(), execPath); err != nil {
			// Try to restore the backup
			_ = os.Rename(bakPath, execPath)
			return fmt.Errorf("failed to copy new executable: %w", err)
		}

		// Remove the backup
		_ = os.Remove(bakPath)
	} else {
		// On Unix-like systems, we can replace the running executable directly
		if err := os.Rename(tempFile.Name(), execPath); err != nil {
			return fmt.Errorf("failed to replace current executable: %w", err)
		}
	}

	return nil
}

// ListAvailableVersions returns a list of all available versions
func (c *Client) ListAvailableVersions() ([]string, error) {
	// This is a simplified implementation that would need to be replaced with
	// actual logic to list versions from the DigitalOcean Space
	// For now, we'll just return a mock list
	return []string{"1.0.0", "1.0.1", "1.1.0"}, nil
}

// getBinaryNameForCurrentPlatform returns the binary name for the current platform
func getBinaryNameForCurrentPlatform() string {
	goos := runtime.GOOS
	arch := runtime.GOARCH

	if goos == "windows" {
		return fmt.Sprintf("devplan-%s-%s.exe", goos, arch)
	}

	return fmt.Sprintf("devplan-%s-%s", goos, arch)
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func(sourceFile *os.File) {
		_ = sourceFile.Close()
	}(sourceFile)

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func(destFile *os.File) {
		_ = destFile.Close()
	}(destFile)

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	return destFile.Sync()
}
