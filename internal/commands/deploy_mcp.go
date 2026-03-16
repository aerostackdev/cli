package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aerostackdev/cli/internal/api"
	"github.com/aerostackdev/cli/internal/credentials"
	"github.com/aerostackdev/cli/internal/devserver"
	"github.com/aerostackdev/cli/internal/modules/mcpconvert"
	"github.com/aerostackdev/cli/internal/printer"
	"github.com/spf13/cobra"
)

// NewDeployMcpCommand creates the 'aerostack deploy mcp' command
func NewDeployMcpCommand() *cobra.Command {
	var environment string

	cmd := &cobra.Command{
		Use:   "mcp [name]",
		Short: "Deploy an MCP server to Aerostack cloud",
		Long: `Deploy an MCP server to Aerostack's infrastructure by name.
Automatically builds the server before deploying.

Examples:
  aerostack deploy mcp github
  aerostack deploy mcp stripe
  aerostack deploy mcp github --env staging`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// 1. Load API Key
			apiKey := credentials.GetAPIKey()
			if apiKey == "" {
				return fmt.Errorf("not logged in. Run 'aerostack login' first")
			}

			// 2. Resolve name/slug and project directory
			var name string
			var projectDir string

			if len(args) > 0 {
				// Name provided: aerostack deploy mcp github
				// Strip "mcp-" prefix if user typed it (e.g. "mcp-github" → "github")
				name = strings.TrimPrefix(args[0], "mcp-")
				// Look for MCP/mcp-{name} relative to cwd, or use cwd if already inside it
				cwd, _ := os.Getwd()
				candidate := filepath.Join(cwd, "MCP", "mcp-"+name)
				if _, err := os.Stat(candidate); err == nil {
					projectDir = candidate
				} else {
					// Maybe we're already inside the repo root or MCP dir
					candidate2 := filepath.Join(cwd, "mcp-"+name)
					if _, err := os.Stat(candidate2); err == nil {
						projectDir = candidate2
					} else {
						return fmt.Errorf("MCP server 'mcp-%s' not found. Expected at MCP/mcp-%s/", name, name)
					}
				}
			} else {
				// No name: infer from current directory (e.g. inside MCP/mcp-github/)
				cwd, _ := os.Getwd()
				base := filepath.Base(cwd)
				name = strings.TrimPrefix(base, "mcp-")
				projectDir = cwd
			}

			// 3. Parse aerostack.toml for metadata (env vars, description, category, tags)
			var envVars []string
			var description, category, tags string
			tomlPath := filepath.Join(projectDir, "aerostack.toml")
			if cfg, err := devserver.ParseAerostackToml(tomlPath); err == nil {
				envVars = cfg.EnvVars
				description = cfg.Description
				category = cfg.Category
				if len(cfg.Tags) > 0 {
					tagsBytes, _ := json.Marshal(cfg.Tags)
					tags = string(tagsBytes)
				}
			}

			// 4. Build
			printer.Step("Building MCP server 'mcp-%s'...", name)
			workerPath, err := mcpconvert.BundleWorker(projectDir)
			if err != nil {
				return fmt.Errorf("build failed: %w", err)
			}

			// 5. Deploy
			printer.Step("Deploying 'mcp-%s' to Aerostack (%s)...", name, environment)
			if len(envVars) > 0 {
				printer.Step("  Required credentials: %s", strings.Join(envVars, ", "))
			}
			resp, err := api.CommunityDeployMcp(apiKey, workerPath, "mcp-"+name, environment, envVars, description, category, tags)
			if err != nil {
				return err
			}

			fmt.Println()
			printer.Success("MCP Server deployed successfully!")
			fmt.Println(printer.KeyVal("Name", "mcp-"+name))
			fmt.Println(printer.KeyVal("URL", resp.WorkerURL))
			fmt.Println(printer.KeyVal("Status", "Hosted"))
			fmt.Println()
			printer.Hint("You can now use this MCP server in your Aerostack agents.")

			return nil
		},
	}

	cmd.Flags().StringVarP(&environment, "env", "e", "production", "Target environment (staging/production)")

	return cmd
}
