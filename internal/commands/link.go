package commands

import (
	"fmt"

	"github.com/aerostackdev/cli/internal/link"
	"github.com/spf13/cobra"
)

// NewLinkCommand creates the 'aerostack link' command
func NewLinkCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "link [project-id]",
		Short: "Link this project to an Aerostack project",
		Long: `Link your local project to an Aerostack project by ID.

Get the project ID from the Aerostack dashboard (Project Settings).
After linking, 'aerostack deploy' will deploy to Aerostack instead of your Cloudflare account.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return linkProject(args[0])
		},
	}

	return cmd
}

func linkProject(projectID string) error {
	if projectID == "" {
		return fmt.Errorf("project ID is required")
	}

	if err := link.Save(projectID); err != nil {
		return fmt.Errorf("save link: %w", err)
	}

	fmt.Printf("✅ Linked to project %s\n", projectID)
	fmt.Println("   Run 'aerostack login' if you haven't, then 'aerostack deploy'")
	return nil
}
