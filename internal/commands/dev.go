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
		Long: `Start the Aerostack local development server with D1, KV, and full local fidelity.

What Aerostack dev gives you (vs raw npx wrangler dev):
  • aerostack.toml as single config — no wrangler.toml to maintain
  • Auto D1 binding — blank projects get env.DB by default
  • --remote staging — debug with real staging data
  • Same stack as deploy — init, dev, deploy share one config

Requires Node.js 18+.

Example:
  aerostack dev                    # Start local dev server
  aerostack dev --port 8787        # Use custom port
  aerostack dev --remote           # Use real Cloudflare bindings`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return startDevServer(port, remote)
		},
	}

	cmd.Flags().IntVarP(&port, "port", "p", 8787, "Port for the dev server")
	cmd.Flags().StringVar(&remote, "remote", "", "Connect to remote environment (staging/production)")

	return cmd
}

func startDevServer(port int, remote string) error {
	fmt.Println("┌─────────────────────────────────────────────────────────┐")
	fmt.Println("│  Aerostack dev  —  One config, D1 included, ready to go  │")
	fmt.Println("└─────────────────────────────────────────────────────────┘")
	fmt.Printf("\n🔧 Starting on http://localhost:%d\n", port)

	// 1. Check for aerostack.toml (fallback to wrangler.toml)
	configPath := "aerostack.toml"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if _, err := os.Stat("wrangler.toml"); err == nil {
			configPath = "wrangler.toml"
			fmt.Println("⚠️  aerostack.toml not found. Falling back to wrangler.toml")
		} else {
			return fmt.Errorf("aerostack.toml not found. Run 'aerostack init' first")
		}
	}

	// 2. Check Node.js (required for D1 via Wrangler/Miniflare)
	nodeVersion, err := devserver.CheckNode()
	if err != nil {
		return err
	}
	fmt.Printf("✓ Node.js %s\n", nodeVersion)

	// 2b. Pre-flight: make sure the target port is free.
	// Wrangler hangs silently if the port is occupied — fail fast instead.
	if err := devserver.CheckPortAvailable(port); err != nil {
		return fmt.Errorf("cannot start dev server: %w", err)
	}

	// 3. Parse config and generate wrangler.toml
	cfg, err := devserver.ParseAerostackToml(configPath)
	if err != nil {
		return fmt.Errorf("failed to parse %s: %w", configPath, err)
	}

	// Ensure at least one D1 binding for local dev (blank template may not have it)
	devserver.EnsureDefaultD1(cfg)
	// Ensure CACHE KV binding (required by SDK)
	devserver.EnsureDefaultKV(cfg)
	// Ensure QUEUE binding (required by SDK)
	devserver.EnsureDefaultQueues(cfg)
	// Ensure AI binding
	devserver.EnsureDefaultAI(cfg)

	// Validate Postgres connection strings
	for _, pg := range cfg.PostgresDatabases {
		if err := devserver.ValidatePostgresConnectionString(pg.ConnectionString); err != nil {
			return fmt.Errorf("invalid Postgres connection for binding '%s': %w", pg.Binding, err)
		}
	}

	// 4. Initialize .aerostack directory for local state
	dotAerostack := ".aerostack"
	if err := os.MkdirAll(dotAerostack, 0755); err != nil {
		return fmt.Errorf("failed to create .aerostack directory: %w", err)
	}

	// 5. Generate wrangler.toml inside .aerostack/ (and per-service configs for multi-worker)
	// This keeps the project root clean — users only see aerostack.toml
	wranglerPath := filepath.Join(dotAerostack, "wrangler.toml")
	if err := devserver.GenerateWranglerToml(cfg, wranglerPath); err != nil {
		return fmt.Errorf("failed to generate wrangler config: %w", err)
	}

	// Multi-worker: generate configs for each service
	workerConfigs := []struct{ name, path string }{{name: "main", path: wranglerPath}}
	for _, svc := range cfg.Services {
		svcPath := filepath.Join(dotAerostack, "wrangler-"+svc.Name+".toml")
		if err := devserver.GenerateWranglerTomlForService(cfg, svc, svcPath); err != nil {
			return fmt.Errorf("failed to generate wrangler config for %s: %w", svc.Name, err)
		}
		workerConfigs = append(workerConfigs, struct{ name, path string }{svc.Name, svcPath})
	}

	// Show database configuration
	dbMsg := fmt.Sprintf("D1: %d", len(cfg.D1Databases))
	if len(cfg.PostgresDatabases) > 0 {
		dbMsg += fmt.Sprintf(", Postgres: %d", len(cfg.PostgresDatabases))
	}
	if len(cfg.Services) > 0 {
		dbMsg += fmt.Sprintf(", Services: %d", len(cfg.Services)+1)
	}
	fmt.Printf("📄 Generated %s (%s)\n", wranglerPath, dbMsg)

	if _, err := os.Stat(".dev.vars"); err == nil {
		fmt.Println("🔐 Loading .dev.vars for local secrets")
	}

	if remote != "" {
		fmt.Printf("🌐 Connected to remote environment: %s\n", remote)
	}

	// 6. Build Hyperdrive env vars for local Postgres
	hyperdriveEnv := make(map[string]string)
	if remote == "" {
		for _, pg := range cfg.PostgresDatabases {
			envKey := "CLOUDFLARE_HYPERDRIVE_LOCAL_CONNECTION_STRING_" + pg.Binding
			hyperdriveEnv[envKey] = pg.ConnectionString
		}
	}

	// 7. Run wrangler dev (single or multi-worker)
	var cmds []*exec.Cmd
	exitChan := make(chan error, len(workerConfigs))
	for i, wc := range workerConfigs {
		p := port + i
		cmd, err := devserver.RunWranglerDev(wc.path, p, remote, hyperdriveEnv)
		if err != nil {
			// Kill any already started
			for _, c := range cmds {
				if c != nil && c.Process != nil {
					devserver.KillProcessGroup(c.Process)
				}
			}
			return err
		}
		cmds = append(cmds, cmd)
		fmt.Printf("   [%s] http://localhost:%d\n", wc.name, p)

		// Monitor process exit
		go func(c *exec.Cmd, name string) {
			err := c.Wait()
			exitChan <- fmt.Errorf("worker [%s] exited: %v", name, err)
		}(cmd, wc.name)
	}

	fmt.Println("\n✅ Dev server ready!")
	fmt.Println("   Press Ctrl+C to stop")

	// Wait for interrupt signal OR process exit
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-sigChan:
		fmt.Printf("\n👋 Received %v. Shutting down dev server...\n", s)
	case err := <-exitChan:
		fmt.Printf("\n❌ %v\n", err)
	}

	for _, c := range cmds {
		if c != nil && c.Process != nil {
			devserver.KillProcessGroup(c.Process)
		}
	}
	return nil
}
