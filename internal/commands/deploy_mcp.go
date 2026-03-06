package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aerostackdev/cli/internal/api"
	"github.com/aerostackdev/cli/internal/credentials"
	"github.com/aerostackdev/cli/internal/printer"
	"github.com/spf13/cobra"
)

// NewDeployMcpCommand creates the 'aerostack deploy mcp' command
func NewDeployMcpCommand() *cobra.Command {
	var environment string
	var slug string

	cmd := &cobra.Command{
		Use:   "mcp [file]",
		Short: "Deploy an MCP server to Aerostack cloud",
		Long: `Deploy a bundled MCP server Worker to Aerostack's infrastructure.
The server will be hosted and accessible via a workers.dev subdomain.

Example:
  aerostack deploy mcp dist/index.js --slug my-mcp-server`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workerPath := "dist/index.js"
			if len(args) > 0 {
				workerPath = args[0]
			}

			if _, err := os.Stat(workerPath); os.IsNotExist(err) {
				return fmt.Errorf("worker file not found: %s. Build your MCP server first", workerPath)
			}

			// 1. Load API Key
			apiKey := credentials.GetAPIKey()
			if apiKey == "" {
				return fmt.Errorf("not logged in. Run 'aerostack login' first")
			}

			// 2. Resolve Slug
			if slug == "" {
				// Try to get from folder name
				cwd, _ := os.Getwd()
				slug = filepath.Base(cwd)
			}

			printer.Step("Deploying MCP server '%s' (%s)...", slug, environment)

			// 3. Deploy
			resp, err := api.CommunityDeployMcp(apiKey, workerPath, slug, environment)
			if err != nil {
				return err
			}

			fmt.Println()
			printer.Success("MCP Server deployed successfully!")
			fmt.Println(printer.KeyVal("URL", resp.WorkerURL))
			fmt.Println(printer.KeyVal("Slug", resp.Slug))
			fmt.Println(printer.KeyVal("Status", "Hosted"))
			fmt.Println()
			printer.Hint("You can now use this MCP server in your Aerostack agents.")

			return nil
		},
	}

	cmd.Flags().StringVarP(&environment, "env", "e", "production", "Target environment (staging/production)")
	cmd.Flags().StringVar(&slug, "slug", "", "The unique slug of the MCP server (required if not in a named directory)")

	return cmd
}
