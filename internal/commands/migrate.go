package commands

import (
	"fmt"
	"os"

	"github.com/aerostackdev/cli/internal/agent"
	"github.com/aerostackdev/cli/internal/modules/migration"
	"github.com/aerostackdev/cli/internal/pkg"
	"github.com/spf13/cobra"
)

func NewMigrateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate a Cloudflare Worker project to Aerostack",
		Long: `Automatically detects a wrangler.toml file, generates an aerostack.toml config,
and uses AI to suggest code updates for compatibility.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, _ := os.Getwd()

			// 1. Detect
			fmt.Println("🔍 Looking for Cloudflare Worker project...")
			wConfig, err := migration.Detect(cwd)
			if err != nil {
				return fmt.Errorf("detection failed: %w", err)
			}
			fmt.Printf("✅ Detected Wrangler project: %s\n", wConfig.Name)

			// 2. Convert Config
			fmt.Println("⚙️  Generating Aerostack configuration...")
			aConfig := migration.ConvertWranglerToAerostack(wConfig)

			// Write aerostack.toml
			if err := migration.GenerateAerostackToml(aConfig, "aerostack.toml"); err != nil {
				return fmt.Errorf("failed to write aerostack.toml: %w", err)
			}
			fmt.Println("✅ Created aerostack.toml")

			// 3. AI Refactor Check
			fmt.Println("🧠 Initializing AI for code compatibility check...")
			store, err := pkg.NewStore(cwd)
			if err != nil {
				// Warn but proceed? Or fail?
				fmt.Printf("⚠️  Warning: Could not init PKG store: %v. AI features limited.\n", err)
			}

			// Initialize Agent (optional - skip AI check if no keys)
			var ag *agent.Agent
			if store != nil {
				ag, err = agent.NewAgent(store, false)
				if err != nil {
					fmt.Printf("⚠️  Skipping AI check (no API key): %v\n", err)
					ag = nil
				}
			}

			if ag != nil {
				refactor := migration.NewRefactorAI(ag)
				// Check main entry file
				entryFile := wConfig.Main
				if entryFile == "" {
					entryFile = "src/index.ts" // default
				}

				if err := refactor.CheckCompatibility(cmd.Context(), entryFile); err != nil {
					fmt.Printf("⚠️  AI check failed: %v\n", err)
				}
			} else {
				fmt.Println("ℹ️  Skipping AI check (Agent not available).")
			}

			fmt.Println("\n🎉 Migration setup complete! Run 'aerostack dev' to test your project.")
			return nil
		},
	}

	return cmd
}
