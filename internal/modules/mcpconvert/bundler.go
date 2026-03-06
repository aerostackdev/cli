package mcpconvert

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// BundleWorker runs esbuild on the generated Worker project and returns the bundled file path.
func BundleWorker(projectDir string) (string, error) {
	// Ensure node_modules exist
	if _, err := os.Stat(filepath.Join(projectDir, "node_modules")); os.IsNotExist(err) {
		cmd := exec.Command("npm", "install")
		cmd.Dir = projectDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("npm install failed: %w", err)
		}
	}

	distDir := filepath.Join(projectDir, "dist")
	os.MkdirAll(distDir, 0755)

	outFile := filepath.Join(distDir, "index.js")
	entryFile := filepath.Join(projectDir, "src", "index.ts")

	// Check if entry is .ts or .js
	if _, err := os.Stat(entryFile); os.IsNotExist(err) {
		entryFile = filepath.Join(projectDir, "src", "index.js")
	}

	cmd := exec.Command("npx", "--yes", "esbuild",
		entryFile,
		"--bundle",
		fmt.Sprintf("--outfile=%s", outFile),
		"--format=esm",
		"--minify",
		"--external:node:*",
		"--external:cloudflare:*",
	)
	cmd.Dir = projectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("esbuild bundle failed: %w", err)
	}

	return outFile, nil
}
