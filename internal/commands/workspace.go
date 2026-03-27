package commands

import (
	"fmt"

	"github.com/aerostackdev/cli/internal/api"
	"github.com/aerostackdev/cli/internal/credentials"
	"github.com/spf13/cobra"
)

// NewWorkspaceCommand creates the 'aerostack workspace' root command.
func NewWorkspaceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workspace",
		Short: "Manage your MCP workspaces",
		Long: `List, create, and switch between workspaces.

A workspace is a private collection of skills and MCP servers exposed through a single gateway URL.
You can have multiple workspaces for different contexts (personal, work, per-client, etc.).

Examples:
  aerostack workspace list
  aerostack workspace create "Work Projects"
  aerostack workspace use my-workspace`,
	}

	cmd.AddCommand(NewWorkspaceListCommand())
	cmd.AddCommand(NewWorkspaceCreateCommand())
	cmd.AddCommand(NewWorkspaceUseCommand())
	cmd.AddCommand(NewWorkspaceTestCommand())
	return cmd
}

// ─── workspace list ───────────────────────────────────────────────────────────

// NewWorkspaceListCommand creates 'aerostack workspace list'.
func NewWorkspaceListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List your workspaces",
		RunE: func(cmd *cobra.Command, args []string) error {
			apiKey := credentials.GetAPIKey()
			if apiKey == "" {
				return fmt.Errorf("not logged in. Run: aerostack login")
			}

			workspaces, err := api.WorkspaceList(apiKey)
			if err != nil {
				return err
			}

			cfg, _ := credentials.LoadConfig()
			activeSlug := ""
			if cfg != nil {
				activeSlug = cfg.ActiveWorkspace
			}

			if len(workspaces) == 0 {
				fmt.Println("No workspaces found. Create one with: aerostack workspace create \"My Workspace\"")
				return nil
			}

			fmt.Printf("%-4s %-24s %-20s %s\n", "", "NAME", "SLUG", "GATEWAY URL")
			fmt.Printf("%-4s %-24s %-20s %s\n", "----", "----", "----", "-----------")
			for _, ws := range workspaces {
				active := "  "
				if ws.Slug == activeSlug {
					active = "* "
				}
				name := ws.Name
				if len(name) > 22 {
					name = name[:19] + "..."
				}
				fmt.Printf("%-4s %-24s %-20s %s\n", active, name, ws.Slug, ws.GatewayURL)
			}
			fmt.Println("\n* = active workspace")
			return nil
		},
	}
}

// ─── workspace use ────────────────────────────────────────────────────────────

// NewWorkspaceUseCommand creates 'aerostack workspace use <slug>'.
func NewWorkspaceUseCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "use <slug>",
		Short: "Set the active workspace",
		Long: `Set the active workspace for subsequent commands.
Skills installed with 'aerostack skill install' go into the active workspace by default.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			targetSlug := args[0]

			apiKey := credentials.GetAPIKey()
			if apiKey == "" {
				return fmt.Errorf("not logged in. Run: aerostack login")
			}

			// Verify the workspace exists
			workspaces, err := api.WorkspaceList(apiKey)
			if err != nil {
				return err
			}

			var found bool
			for _, ws := range workspaces {
				if ws.Slug == targetSlug {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("workspace '%s' not found. Run 'aerostack workspace list' to see available workspaces", targetSlug)
			}

			cfg, err := credentials.LoadConfig()
			if err != nil {
				cfg = &credentials.CLIConfig{}
			}
			cfg.ActiveWorkspace = targetSlug
			if err := credentials.SaveConfig(cfg); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Printf("✓ Active workspace set to '%s'\n", targetSlug)
			fmt.Println("  Skills installed with 'aerostack skill install' will go into this workspace.")
			return nil
		},
	}
}

// ─── workspace create ─────────────────────────────────────────────────────────

// NewWorkspaceCreateCommand creates 'aerostack workspace create <name>'.
func NewWorkspaceCreateCommand() *cobra.Command {
	var setActive bool

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new workspace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			apiKey := credentials.GetAPIKey()
			if apiKey == "" {
				return fmt.Errorf("not logged in. Run: aerostack login")
			}

			fmt.Printf("Creating workspace '%s'...\n", name)
			ws, err := api.WorkspaceCreate(apiKey, name)
			if err != nil {
				return err
			}

			fmt.Printf("✓ Workspace created!\n")
			fmt.Printf("  Slug:    %s\n", ws.Slug)
			fmt.Printf("  Gateway: %s\n", ws.GatewayURL)

			if setActive {
				cfg, _ := credentials.LoadConfig()
				if cfg == nil {
					cfg = &credentials.CLIConfig{}
				}
				cfg.ActiveWorkspace = ws.Slug
				if err := credentials.SaveConfig(cfg); err == nil {
					fmt.Printf("  Set as active workspace.\n")
				}
			} else {
				fmt.Printf("\nSet as active with: aerostack workspace use %s\n", ws.Slug)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&setActive, "use", false, "Set this workspace as active immediately")
	return cmd
}

// ─── workspace test ──────────────────────────────────────────────────────────

// NewWorkspaceTestCommand creates 'aerostack workspace test [slug]'.
func NewWorkspaceTestCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "test [slug]",
		Short: "Test workspace connectivity by listing all available tools",
		Long: `Calls tools/list on the workspace gateway and displays all discovered tools.
If no slug is provided, uses the active workspace.

Examples:
  aerostack workspace test
  aerostack workspace test my-workspace`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiKey := credentials.GetAPIKey()
			if apiKey == "" {
				return fmt.Errorf("not logged in. Run: aerostack login")
			}

			// Resolve workspace: arg or active
			var targetSlug string
			if len(args) > 0 {
				targetSlug = args[0]
			} else {
				cfg, _ := credentials.LoadConfig()
				if cfg != nil && cfg.ActiveWorkspace != "" {
					targetSlug = cfg.ActiveWorkspace
				} else {
					return fmt.Errorf("no workspace specified and no active workspace. Run: aerostack workspace use <slug>")
				}
			}

			// Find workspace ID from slug
			workspaces, err := api.WorkspaceList(apiKey)
			if err != nil {
				return err
			}

			var wsID string
			for _, ws := range workspaces {
				if ws.Slug == targetSlug {
					wsID = ws.ID
					break
				}
			}
			if wsID == "" {
				return fmt.Errorf("workspace '%s' not found", targetSlug)
			}

			fmt.Printf("Testing workspace '%s'...\n\n", targetSlug)

			tools, err := api.WorkspaceTestTools(apiKey, wsID)
			if err != nil {
				return fmt.Errorf("test failed: %w", err)
			}

			if len(tools) == 0 {
				fmt.Println("No tools found. Add MCP servers or skills to this workspace first.")
				return nil
			}

			fmt.Printf("%-36s %-20s %s\n", "TOOL", "SERVER", "DESCRIPTION")
			fmt.Printf("%-36s %-20s %s\n", "----", "------", "-----------")
			for _, t := range tools {
				name := t.Name
				if len(name) > 34 {
					name = name[:31] + "..."
				}
				desc := t.Description
				if len(desc) > 50 {
					desc = desc[:47] + "..."
				}
				server := t.ServerSlug
				if server == "" {
					// Extract from namespaced name
					if idx := len(name); idx > 0 {
						parts := splitOnce(t.Name, "__")
						if len(parts) == 2 {
							server = parts[0]
						}
					}
				}
				if len(server) > 18 {
					server = server[:15] + "..."
				}
				fmt.Printf("%-36s %-20s %s\n", name, server, desc)
			}

			fmt.Printf("\n%d tools available across workspace '%s'\n", len(tools), targetSlug)
			return nil
		},
	}
}

// splitOnce splits s on the first occurrence of sep, returning [before, after].
// If sep is not found, returns [s].
func splitOnce(s, sep string) []string {
	idx := 0
	for i := 0; i <= len(s)-len(sep); i++ {
		if s[i:i+len(sep)] == sep {
			idx = i
			return []string{s[:idx], s[idx+len(sep):]}
		}
	}
	return []string{s}
}
