package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/aerostackdev/cli/internal/devserver"
	"github.com/spf13/cobra"
)

// NewTestCommand creates the 'aerostack test' command
func NewTestCommand() *cobra.Command {
	var coverage bool

	cmd := &cobra.Command{
		Use:   "test",
		Short: "Run tests with Vitest",
		Long: `Run tests using Vitest with Workers runtime support.
Generates wrangler.toml and builds the project before running tests.
Uses @cloudflare/vitest-pool-workers for D1/KV/bindings in tests.

Example:
  aerostack test
  aerostack test --coverage`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTests(coverage)
		},
	}

	cmd.Flags().BoolVar(&coverage, "coverage", false, "Generate coverage report")
	return cmd
}

func runTests(coverage bool) error {
	// 1. Check for aerostack.toml
	if _, err := os.Stat("aerostack.toml"); os.IsNotExist(err) {
		return fmt.Errorf("aerostack.toml not found. Run 'aerostack init' first")
	}

	// 2. Parse aerostack.toml and generate wrangler.toml (required for vitest config)
	cfg, err := devserver.ParseAerostackToml("aerostack.toml")
	if err != nil {
		return fmt.Errorf("failed to parse aerostack.toml: %w", err)
	}
	devserver.EnsureDefaultD1(cfg)

	wranglerPath := filepath.Join(".aerostack", "wrangler.toml")
	if err := devserver.GenerateWranglerToml(cfg, wranglerPath); err != nil {
		return fmt.Errorf("failed to generate wrangler config: %w", err)
	}

	// 3. Run build so dist/worker.js exists (vitest pool uses wrangler config)
	buildCmd := exec.Command("npx", "esbuild", cfg.Main, "--bundle", "--outfile=dist/worker.js", "--format=esm", "--alias:@shared=./shared")
	buildCmd.Dir = "."
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		return fmt.Errorf("build failed (required for tests): %w", err)
	}

	// 4. Run vitest
	vitestArgs := []string{"vitest", "run"}
	if coverage {
		vitestArgs = append(vitestArgs, "--coverage")
	}

	testCmd := exec.Command("npx", vitestArgs...)
	testCmd.Dir = "."
	testCmd.Stdout = os.Stdout
	testCmd.Stderr = os.Stderr
	testCmd.Env = append(os.Environ(), "NPX_UPDATE_NOTIFIER=false")

	// Ensure we're in project root
	absDir, _ := filepath.Abs(".")
	testCmd.Dir = absDir

	if err := testCmd.Run(); err != nil {
		return fmt.Errorf("tests failed: %w", err)
	}
	return nil
}
