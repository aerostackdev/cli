package mcpconvert

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// FetchNPMPackage downloads an npm package and extracts it to a temp directory.
// Returns the path to the extracted package directory.
func FetchNPMPackage(packageName string) (string, error) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "aerostack-mcp-convert-*")
	if err != nil {
		return "", fmt.Errorf("create temp dir: %w", err)
	}

	// Use npm pack to download the tarball
	cmd := exec.Command("npm", "pack", packageName, "--pack-destination", tmpDir)
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("npm pack failed: %s\n%w", string(output), err)
	}

	// Find the tarball
	tgzName := strings.TrimSpace(string(output))
	// npm pack might output with scope prefix and version
	tgzPath := filepath.Join(tmpDir, tgzName)
	if _, err := os.Stat(tgzPath); os.IsNotExist(err) {
		// Try to find any .tgz file
		entries, _ := os.ReadDir(tmpDir)
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), ".tgz") {
				tgzPath = filepath.Join(tmpDir, e.Name())
				break
			}
		}
	}

	// Extract tarball
	extractDir := filepath.Join(tmpDir, "extracted")
	os.MkdirAll(extractDir, 0755)
	cmd = exec.Command("tar", "xzf", tgzPath, "-C", extractDir)
	if err := cmd.Run(); err != nil {
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("extract tarball: %w", err)
	}

	// npm pack extracts to a "package/" subdirectory
	packageDir := filepath.Join(extractDir, "package")
	if _, err := os.Stat(packageDir); os.IsNotExist(err) {
		// Some packages extract differently, check for any subdirectory
		entries, _ := os.ReadDir(extractDir)
		if len(entries) == 1 && entries[0].IsDir() {
			packageDir = filepath.Join(extractDir, entries[0].Name())
		} else {
			packageDir = extractDir
		}
	}

	return packageDir, nil
}

// FetchGitHubRepo clones a GitHub repo to a temp directory.
func FetchGitHubRepo(repoURL string) (string, error) {
	tmpDir, err := os.MkdirTemp("", "aerostack-mcp-convert-*")
	if err != nil {
		return "", fmt.Errorf("create temp dir: %w", err)
	}

	cmd := exec.Command("git", "clone", "--depth=1", repoURL, tmpDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("git clone failed: %s\n%w", string(output), err)
	}

	return tmpDir, nil
}

// FetchNPMPackageInfo fetches package metadata from the npm registry.
func FetchNPMPackageInfo(packageName string) (*NPMPackageInfo, error) {
	url := fmt.Sprintf("https://registry.npmjs.org/%s/latest", packageName)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch npm info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("npm registry returned %d: %s", resp.StatusCode, string(body))
	}

	var info NPMPackageInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("parse npm info: %w", err)
	}
	return &info, nil
}

// NPMPackageInfo holds relevant npm package metadata.
type NPMPackageInfo struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Description string            `json:"description"`
	Main        string            `json:"main"`
	Bin         json.RawMessage   `json:"bin"`
	Keywords    []string          `json:"keywords"`
	Repository  *NPMRepo          `json:"repository"`
}

// NPMRepo is the repository field in package.json.
type NPMRepo struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

// Cleanup removes a temporary directory created by Fetch* functions.
func Cleanup(dir string) {
	if dir != "" && strings.Contains(dir, "aerostack-mcp-convert") {
		os.RemoveAll(dir)
	}
}
