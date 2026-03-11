package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aerostackdev/cli/internal/api"
	"github.com/aerostackdev/cli/internal/credentials"
	"github.com/aerostackdev/cli/internal/printer"
	"github.com/spf13/cobra"
)

// NewDeploySkillCommand creates the 'aerostack deploy skill' command
func NewDeploySkillCommand() *cobra.Command {
	var environment string

	cmd := &cobra.Command{
		Use:   "skill [name]",
		Short: "Deploy a skill to Aerostack cloud",
		Long: `Deploy a skill to Aerostack's infrastructure by name.
Reads the SKILL.md file from the skills directory and publishes it to the registry.

Examples:
  aerostack deploy skill new-mcp
  aerostack deploy skill debug-worker
  aerostack deploy skill new-mcp --env staging`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// 1. Load API Key
			apiKey := credentials.GetAPIKey()
			if apiKey == "" {
				return fmt.Errorf("not logged in. Run 'aerostack login' first")
			}

			// 2. Resolve name and SKILL.md path
			var name string
			var skillPath string

			if len(args) > 0 {
				name = args[0]
				cwd, _ := os.Getwd()

				// Look for skills/{name}/SKILL.md relative to cwd
				candidate := filepath.Join(cwd, "skills", name, "SKILL.md")
				if _, err := os.Stat(candidate); err == nil {
					skillPath = candidate
				} else {
					// Maybe we're already inside the skill dir or skills/
					candidate2 := filepath.Join(cwd, name, "SKILL.md")
					if _, err := os.Stat(candidate2); err == nil {
						skillPath = candidate2
					} else {
						// Maybe SKILL.md is in current dir
						candidate3 := filepath.Join(cwd, "SKILL.md")
						if _, err := os.Stat(candidate3); err == nil {
							skillPath = candidate3
						} else {
							return fmt.Errorf("SKILL.md not found for skill '%s'. Expected at skills/%s/SKILL.md", name, name)
						}
					}
				}
			} else {
				// No name: look for SKILL.md in current directory
				cwd, _ := os.Getwd()
				candidate := filepath.Join(cwd, "SKILL.md")
				if _, err := os.Stat(candidate); err != nil {
					return fmt.Errorf("SKILL.md not found in current directory. Provide a skill name or run from a skill directory")
				}
				skillPath = candidate
				name = filepath.Base(cwd)
			}

			// Strip common prefixes if user typed full dir name
			name = strings.TrimPrefix(name, "skill-")

			// 3. Read SKILL.md
			printer.Step("Reading SKILL.md for '%s'...", name)
			content, err := os.ReadFile(skillPath)
			if err != nil {
				return fmt.Errorf("failed to read SKILL.md: %w", err)
			}
			if len(strings.TrimSpace(string(content))) == 0 {
				return fmt.Errorf("SKILL.md is empty")
			}

			// 4. Deploy
			printer.Step("Deploying skill '%s' to Aerostack (%s)...", name, environment)
			resp, err := api.CommunityDeploySkill(apiKey, name, string(content), environment)
			if err != nil {
				return err
			}

			fmt.Println()
			printer.Success("Skill deployed successfully!")
			fmt.Println(printer.KeyVal("Name", name))
			fmt.Println(printer.KeyVal("URL", resp.URL))
			fmt.Println(printer.KeyVal("Status", "Published"))
			fmt.Println()
			printer.Hint("Your skill is now available in the Aerostack marketplace.")

			return nil
		},
	}

	cmd.Flags().StringVarP(&environment, "env", "e", "production", "Target environment (staging/production)")

	return cmd
}
