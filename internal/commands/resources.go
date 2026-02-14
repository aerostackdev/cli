package commands

import (
	"fmt"
	"os"

	"github.com/aerostackdev/cli/internal/devserver"
	"github.com/aerostackdev/cli/internal/provision"
	"github.com/spf13/cobra"
)

// NewResourcesCommand creates the 'aerostack resources' command
func NewResourcesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resources",
		Short: "Provision Cloudflare resources in your account",
		Long: `Create D1, KV, R2, Queues (etc.) in your Cloudflare account based on aerostack.toml.
Use with --cloudflare before deploy, or run standalone to set up resources.

Requires: npx wrangler login (or CLOUDFLARE_API_TOKEN)

Examples:
  aerostack resources create --env staging
  aerostack resources create --env production`,
	}

	cmd.AddCommand(newResourcesCreateCommand())
	return cmd
}

func newResourcesCreateCommand() *cobra.Command {
	var env string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create resources (D1, KV, etc.) in your Cloudflare account",
		Long: `Auto-provisions resources from aerostack.toml when IDs are placeholders.
Updates aerostack.toml with real IDs. Run before deploy --cloudflare or let deploy auto-run this.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runResourcesCreate(env)
		},
	}
	cmd.Flags().StringVarP(&env, "env", "e", "staging", "Target environment (staging/production)")
	return cmd
}

func runResourcesCreate(env string) error {
	if env != "staging" && env != "production" {
		return fmt.Errorf("invalid env %q: use staging or production", env)
	}

	if _, err := os.Stat("aerostack.toml"); os.IsNotExist(err) {
		return fmt.Errorf("aerostack.toml not found. Run 'aerostack init' first")
	}

	cfg, err := devserver.ParseAerostackToml("aerostack.toml")
	if err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	projectRoot, _ := os.Getwd()
	if projectRoot == "" {
		projectRoot = "."
	}

	fmt.Printf("🔍 Provisioning resources for %s...\n", env)
	if err := provision.ProvisionCloudflareResources(cfg, env, projectRoot); err != nil {
		return err
	}
	fmt.Println("\n✅ Resources ready. Run 'aerostack deploy --cloudflare' to deploy.")
	return nil
}
