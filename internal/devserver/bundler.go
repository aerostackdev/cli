package devserver

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/evanw/esbuild/pkg/api"
)

// Bundle transpiles and bundles the user's TypeScript/JavaScript code into a single
// ESM module that can be executed by workerd.
func Bundle(entryPoint, dotAerostack string) (string, error) {
	distDir := filepath.Join(dotAerostack, "dist")
	if err := os.MkdirAll(distDir, 0755); err != nil {
		return "", err
	}

	outputPath := filepath.Join(distDir, "index.js")

	result := api.Build(api.BuildOptions{
		EntryPoints: []string{entryPoint},
		Outfile:     outputPath,
		Bundle:      true,
		Format:      api.FormatESModule,
		Target:      api.ESNext,
		Platform:    api.PlatformBrowser, // Cloudflare Workers are browser-like
		Write:       true,
		LogLevel:    api.LogLevelInfo,
		External:    []string{"node:*"}, // Cloudflare Workers don't support Node built-ins directly
	})

	if len(result.Errors) > 0 {
		for _, err := range result.Errors {
			fmt.Printf("❌ esbuild error: %s\n", err.Text)
		}
		return "", fmt.Errorf("bundling failed with %d errors", len(result.Errors))
	}

	fmt.Printf("📦 Code bundled successfully: %s\n", outputPath)
	return outputPath, nil
}
