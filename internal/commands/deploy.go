package commands

import (
	"fmt"
	"os"

	"github.com/aerostackdev/cli/internal/devserver"
	"github.com/spf13/cobra"
)

// NewDeployCommand creates the 'aerostack deploy' command
func NewDeployCommand() *cobra.Command {
	var environment string
	var allServices bool

	cmd := &cobra.Command{
		Use:   "deploy [service-name]",
		Short: "Deploy services to Aerostack cloud",
		Long: `Deploy your services to the Aerostack cloud infrastructure.

Supports multi-environment deployments (staging, production).
Uses wrangler under the hood — ensure you're logged in (wrangler login or aerostack login).

Example:
  aerostack deploy --env staging
  aerostack deploy --env production
  aerostack deploy --all --env staging`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var serviceName string
			if len(args) > 0 {
				serviceName = args[0]
			}
			return deployService(serviceName, environment, allServices)
		},
	}

	cmd.Flags().StringVarP(&environment, "env", "e", "staging", "Target environment (staging/production)")
	cmd.Flags().BoolVar(&allServices, "all", false, "Deploy all services")

	return cmd
}

func deployService(service, env string, all bool) error {
	// Validate env
	if env != "staging" && env != "production" {
		return fmt.Errorf("invalid env %q: use staging or production", env)
	}

	// 1. Check aerostack.toml
	if _, err := os.Stat("aerostack.toml"); os.IsNotExist(err) {
		return fmt.Errorf("aerostack.toml not found. Run 'aerostack init' first")
	}

	// 2. Check Node.js (wrangler needs it)
	if _, err := devserver.CheckNode(); err != nil {
		return err
	}

	// 3. Parse config and generate wrangler.toml
	cfg, err := devserver.ParseAerostackToml("aerostack.toml")
	if err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}
	devserver.EnsureDefaultD1(cfg)

	wranglerPath := "wrangler.toml"
	if err := devserver.GenerateWranglerToml(cfg, wranglerPath); err != nil {
		return fmt.Errorf("failed to generate wrangler.toml: %w", err)
	}

	// 4. Deploy (single service for now; --all not yet implemented)
	if all {
		fmt.Printf("🚀 Deploying all services to %s...\n", env)
		// TODO: multi-service deploy
	}
	fmt.Printf("🚀 Deploying to %s...\n", env)

	if err := devserver.RunWranglerDeploy(wranglerPath, env); err != nil {
		return err
	}

	fmt.Println("\n✅ Deployment successful!")
	return nil
}
