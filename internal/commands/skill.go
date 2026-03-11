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

// NewSkillCommand creates the 'aerostack skill' root command.
func NewSkillCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "skill",
		Short: "Manage skills in the Aerostack marketplace",
		Long: `Install, publish, list, and remove AI skills from the Aerostack marketplace.

Skills are atomic, single-purpose tools that any LLM can call through your workspace gateway.
Once installed, skills appear automatically through your existing gateway URL — no config changes needed.

Examples:
  aerostack skill install johndoe/github-skill
  aerostack skill install @acme/internal-skill
  aerostack skill publish --name "My Skill" --function abc-123
  aerostack skill list`,
	}

	cmd.AddCommand(NewSkillInstallCommand())
	cmd.AddCommand(NewSkillPublishCommand())
	cmd.AddCommand(NewSkillListCommand())
	cmd.AddCommand(NewSkillRemoveCommand())
	cmd.AddCommand(NewSkillInitCommand())
	cmd.AddCommand(NewSkillPullCommand())
	return cmd
}

// ─── skill init ───────────────────────────────────────────────────────────────

// NewSkillInitCommand creates 'aerostack skill init <name>'.
func NewSkillInitCommand() *cobra.Command {
	var withFunction bool

	cmd := &cobra.Command{
		Use:   "init <name>",
		Short: "Scaffold a new skill",
		Long: `Scaffold a new skill directory.

Two templates available:
  static (default)  — SKILL.md only (AI behavior definition, no custom code)
  function-backed   — SKILL.md + src/index.ts (custom logic deployed as a function)

Examples:
  aerostack skill init github-pr-review
  aerostack skill init daily-digest --function`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := strings.ToLower(strings.TrimSpace(args[0]))
			name = strings.TrimPrefix(name, "skill-")

			printer.Header("Skill Init")

			skillDir := filepath.Join("skills", name)
			if err := os.MkdirAll(skillDir, 0755); err != nil {
				return fmt.Errorf("create directory: %w", err)
			}

			title := strings.ReplaceAll(name, "-", " ")
			title = strings.Title(title) //nolint:staticcheck

			skillMd := fmt.Sprintf(`# %s Skill

## Description
Brief description of what this skill does.

## Trigger Patterns
- When user asks to...
- Triggered by...

## Behavior
Step-by-step what the skill does.

## Examples
- Example usage 1
- Example usage 2

## Configuration
- No configuration required
`, title)

			if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(skillMd), 0644); err != nil {
				return fmt.Errorf("write SKILL.md: %w", err)
			}

			if withFunction {
				srcDir := filepath.Join(skillDir, "src")
				if err := os.MkdirAll(srcDir, 0755); err != nil {
					return fmt.Errorf("create src directory: %w", err)
				}

				indexTs := fmt.Sprintf(`/**
 * %s Skill — Function Logic
 *
 * Called when the skill is invoked. Request body contains
 * the user's configured secrets and any event payload.
 */
export default {
    async fetch(request: Request): Promise<Response> {
        const body = await request.json().catch(() => ({})) as Record<string, unknown>;

        // TODO: implement your skill logic here
        // const secrets = body.secrets as Record<string, string>;
        // const payload = body.payload;

        return Response.json({ success: true, result: 'Hello from %s!' });
    },
};
`, title, name)

				if err := os.WriteFile(filepath.Join(srcDir, "index.ts"), []byte(indexTs), 0644); err != nil {
					return fmt.Errorf("write src/index.ts: %w", err)
				}

				aerostackJson := fmt.Sprintf(`{
    "name": "%s",
    "version": "1.0.0",
    "description": "%s skill",
    "category": "skills",
    "tags": []
}
`, name, title)
				if err := os.WriteFile(filepath.Join(skillDir, "aerostack.json"), []byte(aerostackJson), 0644); err != nil {
					return fmt.Errorf("write aerostack.json: %w", err)
				}

				fmt.Println()
				printer.Success("Function-backed skill scaffolded at skills/%s/", name)
				fmt.Printf("  skills/%s/SKILL.md       ← marketplace listing\n", name)
				fmt.Printf("  skills/%s/src/index.ts   ← function logic\n", name)
				fmt.Printf("  skills/%s/aerostack.json ← metadata\n", name)
				fmt.Println()
				printer.Hint("Next steps:")
				fmt.Printf("  1. Edit skills/%s/SKILL.md\n", name)
				fmt.Printf("  2. Write your logic in skills/%s/src/index.ts\n", name)
				fmt.Printf("  3. aerostack deploy skill %s\n", name)
			} else {
				fmt.Println()
				printer.Success("Static skill scaffolded at skills/%s/SKILL.md", name)
				fmt.Println()
				printer.Hint("Next steps:")
				fmt.Printf("  1. Edit skills/%s/SKILL.md\n", name)
				fmt.Printf("  2. aerostack deploy skill %s\n", name)
				fmt.Println()
				printer.Hint("Want custom function logic? Re-run with --function flag.")
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&withFunction, "function", false, "Scaffold a function-backed skill (SKILL.md + src/index.ts)")
	return cmd
}

// ─── skill pull ───────────────────────────────────────────────────────────────

// NewSkillPullCommand creates 'aerostack skill pull <slug>'.
func NewSkillPullCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "pull <slug>",
		Short: "Pull a skill from Aerostack to your local directory",
		Long: `Downloads a skill's SKILL.md from Aerostack and writes it locally.

Examples:
  aerostack skill pull github-pr-review          (pulls your own skill)
  aerostack skill pull @johndoe/github-pr-review (pulls another developer's skill)`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			slug := args[0]

			apiKey := credentials.GetAPIKey()
			if apiKey == "" {
				return fmt.Errorf("not logged in. Run: aerostack login")
			}

			scopedSlug := slug
			if !strings.HasPrefix(scopedSlug, "@") {
				baseName := strings.TrimPrefix(scopedSlug, "skill-")
				scopedSlug = "skill-" + baseName
			}

			printer.Step("Pulling %s...", scopedSlug)

			resp, err := api.SkillPull(apiKey, scopedSlug)
			if err != nil {
				return err
			}

			// Derive local name from slug: @username/skill-name → name
			baseName := resp.Slug
			if idx := strings.LastIndex(baseName, "/"); idx >= 0 {
				baseName = baseName[idx+1:]
			}
			baseName = strings.TrimPrefix(baseName, "skill-")

			skillDir := filepath.Join("skills", baseName)
			if err := os.MkdirAll(skillDir, 0755); err != nil {
				return fmt.Errorf("create directory: %w", err)
			}

			dest := filepath.Join(skillDir, "SKILL.md")
			if err := os.WriteFile(dest, []byte(resp.Content), 0644); err != nil {
				return fmt.Errorf("write SKILL.md: %w", err)
			}

			fmt.Println()
			printer.Success("Pulled to %s", dest)
			printer.Hint("Edit the file, then run: aerostack deploy skill %s", baseName)
			return nil
		},
	}
}

// ─── skill install ────────────────────────────────────────────────────────────

// NewSkillInstallCommand creates 'aerostack skill install <username/slug>'.
func NewSkillInstallCommand() *cobra.Command {
	var workspaceSlug string

	cmd := &cobra.Command{
		Use:   "install <username/slug>",
		Short: "Install a skill into your active workspace",
		Long: `Install a skill from the Aerostack marketplace into your workspace.

The skill's tools will be available immediately through your gateway URL with no config changes.
Tool names are namespaced to prevent collisions: {skill-slug}__{tool-name}.

Examples:
  aerostack skill install johndoe/github-skill
  aerostack skill install @acme/internal-skill
  aerostack skill install johndoe/slack-skill --workspace my-other-workspace`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref := strings.TrimPrefix(args[0], "@")
			parts := strings.SplitN(ref, "/", 2)
			if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
				return fmt.Errorf("invalid skill reference. Use: username/slug or @username/slug")
			}
			username, slug := parts[0], parts[1]

			apiKey := credentials.GetAPIKey()
			if apiKey == "" {
				return fmt.Errorf("not logged in. Run: aerostack login")
			}

			// 1. Fetch skill metadata
			fmt.Printf("🔍 Looking up %s/%s...\n", username, slug)
			skill, err := api.SkillGet(username, slug)
			if err != nil {
				return err
			}

			// 2. Check team membership if skill is private/team-only
			if skill.Visibility == "team" {
				fmt.Printf("🔒 %s/%s is a team skill — checking membership...\n", username, slug)
				isMember, err := api.TeamCheckMembership(apiKey, username)
				if err != nil {
					return fmt.Errorf("team membership check failed: %w", err)
				}
				if !isMember {
					return fmt.Errorf("access denied: you are not a member of @%s's team.\nAsk %s to invite you via hub.aerostack.ai", username, username)
				}
			} else if skill.Visibility == "private" {
				return fmt.Errorf("'%s/%s' is private and can only be added by its owner", username, slug)
			}

			// 3. Resolve workspace
			ws, err := resolveOrCreateWorkspace(apiKey, workspaceSlug)
			if err != nil {
				return err
			}

			// 4. Add skill to workspace
			fmt.Printf("📥 Installing %s/%s into workspace '%s'...\n", username, slug, ws.Slug)
			if err := api.WorkspaceAddServer(apiKey, ws.ID, skill.ID); err != nil {
				return err
			}

			// 5. Print tool names
			fmt.Printf("\n✓ Skill installed! The following tools are now available in your gateway:\n")
			for _, t := range skill.Tools {
				fmt.Printf("  %s__%s\n", slug, t.Name)
			}
			fmt.Printf("\nGateway URL (unchanged): %s\n", ws.GatewayURL)
			return nil
		},
	}

	cmd.Flags().StringVar(&workspaceSlug, "workspace", "", "Target workspace slug (defaults to active workspace)")
	return cmd
}

// ─── skill publish ────────────────────────────────────────────────────────────

// NewSkillPublishCommand creates 'aerostack skill publish'.
func NewSkillPublishCommand() *cobra.Command {
	var (
		name        string
		description string
		functionID  string
		workerURL   string
		tags        string
		visibility  string
		publish     bool
	)

	cmd := &cobra.Command{
		Use:   "publish",
		Short: "Publish a skill to the Aerostack marketplace",
		Long: `Create or update a skill in the Aerostack marketplace.

A skill can be backed by a deployed Aerostack function (--function) or an external HTTPS server (--worker-url).

Examples:
  # Function-backed skill (no server needed):
  aerostack skill publish --name "Invoice Extractor" --function abc-123 --publish

  # External server:
  aerostack skill publish --name "My Tool" --worker-url https://my-server.example.com

  # Private team skill:
  aerostack skill publish --name "Internal Tool" --function abc-123 --visibility team`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return fmt.Errorf("--name is required")
			}
			if functionID == "" && workerURL == "" {
				return fmt.Errorf("either --function or --worker-url is required")
			}

			apiKey := credentials.GetAPIKey()
			if apiKey == "" {
				return fmt.Errorf("not logged in. Run: aerostack login")
			}

			if visibility == "" {
				visibility = "public"
			}

			payload := api.SkillPublishPayload{
				Name:               name,
				Description:        description,
				BackedByFunctionID: functionID,
				WorkerURL:          workerURL,
				Visibility:         visibility,
				Publish:            publish,
			}
			if tags != "" {
				for _, t := range strings.Split(tags, ",") {
					payload.Tags = append(payload.Tags, strings.TrimSpace(t))
				}
			}

			fmt.Printf("📤 Publishing skill '%s'...\n", name)
			result, err := api.SkillPublish(apiKey, payload)
			if err != nil {
				return err
			}

			fmt.Printf("✓ Skill '%s' created as %s (id: %s)\n", result.Slug, result.Status, result.ID)
			if result.Status == "published" {
				fmt.Printf("  Marketplace: hub.aerostack.ai/skills/%s\n", result.Slug)
			} else {
				fmt.Printf("  Run with --publish to make it live on the marketplace.\n")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Display name for the skill (required)")
	cmd.Flags().StringVar(&description, "description", "", "Short description shown in search results")
	cmd.Flags().StringVar(&functionID, "function", "", "ID of a deployed Aerostack function to back this skill")
	cmd.Flags().StringVar(&workerURL, "worker-url", "", "External HTTPS MCP server URL")
	cmd.Flags().StringVar(&tags, "tags", "", "Comma-separated tags for discovery")
	cmd.Flags().StringVar(&visibility, "visibility", "public", "Visibility: public, team, or private")
	cmd.Flags().BoolVar(&publish, "publish", false, "Publish immediately to the marketplace")
	return cmd
}

// ─── skill list ───────────────────────────────────────────────────────────────

// NewSkillListCommand creates 'aerostack skill list'.
func NewSkillListCommand() *cobra.Command {
	var workspaceSlug string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List skills installed in your workspace",
		RunE: func(cmd *cobra.Command, args []string) error {
			apiKey := credentials.GetAPIKey()
			if apiKey == "" {
				return fmt.Errorf("not logged in. Run: aerostack login")
			}

			ws, err := resolveOrCreateWorkspace(apiKey, workspaceSlug)
			if err != nil {
				return err
			}

			detail, err := api.WorkspaceGet(apiKey, ws.ID)
			if err != nil {
				return fmt.Errorf("failed to fetch workspace details: %w", err)
			}

			if len(detail.Servers) == 0 {
				fmt.Printf("No skills installed in workspace '%s'.\n", ws.Slug)
				fmt.Printf("Install one with: aerostack skill install <username/slug>\n")
				return nil
			}

			fmt.Printf("Skills in workspace '%s' (%s):\n\n", ws.Slug, ws.GatewayURL)
			fmt.Printf("  %-24s %-8s %s\n", "NAME", "TOOLS", "STATUS")
			fmt.Printf("  %-24s %-8s %s\n", "----", "-----", "------")
			for _, s := range detail.Servers {
				status := "enabled"
				if !s.Enabled {
					status = "disabled"
				}
				name := s.Name
				if len(name) > 22 {
					name = name[:19] + "..."
				}
				fmt.Printf("  %-24s %-8d %s\n", name, s.ToolCount, status)
			}
			fmt.Printf("\nGateway: %s\n", ws.GatewayURL)
			return nil
		},
	}

	cmd.Flags().StringVar(&workspaceSlug, "workspace", "", "Workspace slug (defaults to active workspace)")
	return cmd
}

// ─── skill remove ─────────────────────────────────────────────────────────────

// NewSkillRemoveCommand creates 'aerostack skill remove <username/slug>'.
func NewSkillRemoveCommand() *cobra.Command {
	var workspaceSlug string

	cmd := &cobra.Command{
		Use:   "remove <username/slug>",
		Short: "Remove a skill from your workspace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("ℹ  To remove a skill, go to hub.aerostack.ai → Workspaces → %s → remove the skill.\n", workspaceSlug)
			fmt.Println("CLI removal coming in a future release.")
			return nil
		},
	}

	cmd.Flags().StringVar(&workspaceSlug, "workspace", "", "Workspace slug (defaults to active workspace)")
	return cmd
}

// ─── Shared helpers ───────────────────────────────────────────────────────────

// resolveOrCreateWorkspace finds the target workspace: --workspace flag, active workspace, or creates default.
func resolveOrCreateWorkspace(apiKey, slugOverride string) (*api.Workspace, error) {
	workspaces, err := api.WorkspaceList(apiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to list workspaces: %w", err)
	}

	// If --workspace flag is set, find it
	if slugOverride != "" {
		for i, ws := range workspaces {
			if ws.Slug == slugOverride || ws.Name == slugOverride {
				return &workspaces[i], nil
			}
		}
		return nil, fmt.Errorf("workspace '%s' not found. Run 'aerostack workspace list' to see available workspaces", slugOverride)
	}

	// Check active workspace from config
	cfg, _ := credentials.LoadConfig()
	if cfg != nil && cfg.ActiveWorkspace != "" {
		for i, ws := range workspaces {
			if ws.Slug == cfg.ActiveWorkspace {
				return &workspaces[i], nil
			}
		}
	}

	// Use first workspace if only one exists
	if len(workspaces) == 1 {
		return &workspaces[0], nil
	}

	// Multiple workspaces but none active — prompt user to set one
	if len(workspaces) > 1 {
		fmt.Println("You have multiple workspaces. Set an active one with:")
		for _, ws := range workspaces {
			fmt.Printf("  aerostack workspace use %s\n", ws.Slug)
		}
		return nil, fmt.Errorf("no active workspace set")
	}

	// No workspaces exist — create a default one
	fmt.Println("No workspaces found. Creating a default workspace...")
	ws, err := api.WorkspaceCreate(apiKey, "default")
	if err != nil {
		return nil, fmt.Errorf("failed to create default workspace: %w", err)
	}

	// Save it as active
	cfg = &credentials.CLIConfig{ActiveWorkspace: ws.Slug}
	_ = credentials.SaveConfig(cfg)

	fmt.Printf("✓ Created workspace '%s'\n", ws.Slug)
	return ws, nil
}
