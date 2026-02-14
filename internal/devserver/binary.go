package devserver

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

const (
	workerdVersion = "1.20260214.0" // Latest as of today
	workerdBaseURL = "https://github.com/cloudflare/workerd/releases/download/v%s/workerd-%s-%s.gz"
)

// EnsureBinary checks if the workerd binary exists in the .aerostack directory,
// if not, it downloads and extracts it.
func EnsureBinary(dotAerostack string) (string, error) {
	binDir := filepath.Join(dotAerostack, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return "", err
	}

	binaryPath := filepath.Join(binDir, "workerd")
	if _, err := os.Stat(binaryPath); err == nil {
		return binaryPath, nil
	}

	fmt.Printf("📥 workerd binary not found. Downloading v%s...\n", workerdVersion)

	osName := runtime.GOOS
	arch := runtime.GOARCH

	var releaseOS, releaseArch string
	switch osName {
	case "linux":
		releaseOS = "linux"
	case "darwin":
		releaseOS = "darwin"
	default:
		return "", fmt.Errorf("unsupported operating system: %s", osName)
	}

	switch arch {
	case "amd64":
		releaseArch = "64"
	case "arm64":
		releaseArch = "arm64"
	default:
		return "", fmt.Errorf("unsupported architecture: %s", arch)
	}

	downloadURL := fmt.Sprintf(workerdBaseURL, workerdVersion, releaseOS, releaseArch)
	fmt.Printf("🌐 Downloading from %s\n", downloadURL)

	resp, err := http.Get(downloadURL)
	if err != nil {
		return "", fmt.Errorf("failed to download workerd: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download workerd: status %s", resp.Status)
	}

	// Extract .gz (it's a single binary file)
	gr, err := gzip.NewReader(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gr.Close()

	outFile, err := os.OpenFile(binaryPath, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create binary file: %w", err)
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, gr); err != nil {
		return "", fmt.Errorf("failed to extract binary: %w", err)
	}

	fmt.Println("✅ workerd binary installed successfully!")
	return binaryPath, nil
}
