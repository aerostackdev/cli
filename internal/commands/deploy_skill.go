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
		Long: `Deploy a skill to Aerostack's infrastructure.

For static skills (SKILL.md only): publishes to the marketplace.
For function-backed skills (SKILL.md + src/index.ts): auto-deploys the function
to your project, links it to the skill, and publishes both to the marketplace.

Examples:
  aerostack deploy skill daily-digest
  aerostack deploy skill pr-gate --env staging`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// 1. Load API Key
			apiKey := credentials.GetAPIKey()
			if apiKey == "" {
				return fmt.Errorf("not logged in. Run 'aerostack login' first")
			}

			// 2. Resolve skill name and directory
			var name string
			var skillDir string

			cwd, _ := os.Getwd()

			if len(args) > 0 {
				name = args[0]
				// Look for skills/{name}/ relative to cwd
				candidate := filepath.Join(cwd, "skills", name)
				if _, err := os.Stat(filepath.Join(candidate, "SKILL.md")); err == nil {
					skillDir = candidate
				} else {
					// Maybe we're inside the skill dir or skills/
					candidate2 := filepath.Join(cwd, name)
					if _, err := os.Stat(filepath.Join(candidate2, "SKILL.md")); err == nil {
						skillDir = candidate2
					} else if _, err := os.Stat(filepath.Join(cwd, "SKILL.md")); err == nil {
						skillDir = cwd
					} else {
						return fmt.Errorf("SKILL.md not found for skill '%s'. Expected at skills/%s/SKILL.md", name, name)
					}
				}
			} else {
				// No name: use current directory
				if _, err := os.Stat(filepath.Join(cwd, "SKILL.md")); err != nil {
					return fmt.Errorf("SKILL.md not found in current directory. Provide a skill name or run from a skill directory")
				}
				skillDir = cwd
				name = filepath.Base(cwd)
			}

			name = strings.TrimPrefix(name, "skill-")

			// 3. Read SKILL.md
			skillMdPath := filepath.Join(skillDir, "SKILL.md")
			printer.Step("Reading SKILL.md for '%s'...", name)
			skillContent, err := os.ReadFile(skillMdPath)
			if err != nil {
				return fmt.Errorf("failed to read SKILL.md: %w", err)
			}
			if len(strings.TrimSpace(string(skillContent))) == 0 {
				return fmt.Errorf("SKILL.md is empty")
			}

			// 4. Detect if function-backed (src/index.ts exists)
			functionCode := ""
			srcPath := filepath.Join(skillDir, "src", "index.ts")
			if _, err := os.Stat(srcPath); err == nil {
				printer.Step("Detected function-backed skill — reading src/index.ts...")
				code, err := os.ReadFile(srcPath)
				if err != nil {
					return fmt.Errorf("failed to read src/index.ts: %w", err)
				}
				functionCode = string(code)
				printer.Hint("Function code loaded (%d bytes)", len(functionCode))
			}

			// 5. Deploy
			if functionCode != "" {
				printer.Step("Deploying function + skill '%s' to Aerostack (%s)...", name, environment)
			} else {
				printer.Step("Deploying skill '%s' to Aerostack (%s)...", name, environment)
			}

			resp, err := api.CommunityDeploySkill(apiKey, name, string(skillContent), functionCode, environment)
			if err != nil {
				return err
			}

			fmt.Println()
			if resp.FunctionURL != "" {
				printer.Success("Function deployed!")
				fmt.Println(printer.KeyVal("Function URL", resp.FunctionURL))
			}
			printer.Success("Skill published to marketplace!")
			fmt.Println(printer.KeyVal("Skill URL", resp.URL))
			fmt.Println(printer.KeyVal("Slug", resp.Slug))
			fmt.Println()
			if resp.FunctionURL != "" {
				printer.Hint("Function is linked to skill — users who add this skill will invoke your function.")
			} else {
				printer.Hint("Your skill is now available in the Aerostack marketplace.")
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&environment, "env", "e", "production", "Target environment (staging/production)")
	return cmd
}
