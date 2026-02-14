package commands

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/aerostackdev/cli/internal/devserver"
	"github.com/spf13/cobra"
)

// NewDevCommand creates the 'aerostack dev' command
func NewDevCommand() *cobra.Command {
	var port int
	var remote string

	cmd := &cobra.Command{
		Use:   "dev",
		Short: "Start local development server",
		Long: `Start the Aerostack local development server with hot reload.

The dev server runs your services locally using workerd (Cloudflare's runtime),
providing full local fidelity with production. All logs are unified and color-coded.

Example:
  aerostack dev                    # Start local dev server
  aerostack dev --port 8787        # Use custom port
  aerostack dev --remote staging   # Connect to staging environment`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return startDevServer(port, remote)
		},
	}

	cmd.Flags().IntVarP(&port, "port", "p", 8787, "Port for the dev server")
	cmd.Flags().StringVar(&remote, "remote", "", "Connect to remote environment (staging/production)")

	return cmd
}

func startDevServer(port int, remote string) error {
	fmt.Printf("🔧 Starting Aerostack dev server on http://localhost:%d\n", port)

	// 1. Check for aerostack.toml
	if _, err := os.Stat("aerostack.toml"); os.IsNotExist(err) {
		return fmt.Errorf("aerostack.toml not found. Run 'aerostack init' first")
	}

	// 2. Initialize .aerostack directory for local state
	dotAerostack := ".aerostack"
	if err := os.MkdirAll(dotAerostack, 0755); err != nil {
		return fmt.Errorf("failed to create .aerostack directory: %w", err)
	}

	// 3. Initialize D1 Sandbox (SQLite)
	dbDir := filepath.Join(dotAerostack, "data")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	dbPath := filepath.Join(dbDir, "db.sqlite")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		fmt.Println("📦 Initializing local D1 sandbox (SQLite)...")
		file, err := os.Create(dbPath)
		if err != nil {
			return fmt.Errorf("failed to create local database: %w", err)
		}
		file.Close()
	}

	// 4. Download workerd binary if needed
	workerdPath, err := devserver.EnsureBinary(dotAerostack)
	if err != nil {
		return fmt.Errorf("failed to ensure workerd binary: %w", err)
	}

	// 5. Bundle TypeScript code
	_, err = devserver.Bundle("src/index.ts", dotAerostack)
	if err != nil {
		return fmt.Errorf("failed to bundle code: %w", err)
	}

	// 6. Generate workerd config (.capnp)
	// Relative paths from the config file's directory (.aerostack/)
	relBundlePath := "dist/index.js"
	relDBPath := "data"

	configPath, err := devserver.GenerateConfig(dotAerostack, devserver.ConfigData{
		Port:              port,
		BundlePath:        relBundlePath,
		DBPath:            relDBPath,
		CompatibilityDate: "2024-01-01",
	})
	if err != nil {
		return fmt.Errorf("failed to generate workerd config: %w", err)
	}

	if remote != "" {
		fmt.Printf("🌐 Connected to remote environment: %s\n", remote)
	}

	// 7. Start workerd process
	// workerd serve requires: config-file, config-constant-name (paths in embed are relative to config file)
	absWorkerdPath, _ := filepath.Abs(workerdPath)
	absConfigPath, _ := filepath.Abs(configPath)

	cmd := exec.Command(absWorkerdPath, "serve", absConfigPath, "config")
	cmd.Dir = filepath.Dir(absConfigPath) // Resolve embed paths relative to .aerostack/
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start workerd: %w", err)
	}

	fmt.Println("\n✅ Dev server ready!")
	fmt.Printf("   Listening on http://localhost:%d\n", port)
	fmt.Println("   Press Ctrl+C to stop")

	// Wait for interrupt signal to gracefully shut down
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\n👋 Shutting down dev server...")
	if cmd.Process != nil {
		cmd.Process.Kill()
	}
	return nil
}
