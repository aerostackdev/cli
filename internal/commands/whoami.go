package commands

import (
	"fmt"

	"github.com/aerostackdev/cli/internal/api"
	"github.com/aerostackdev/cli/internal/credentials"
	"github.com/aerostackdev/cli/internal/link"
	"github.com/spf13/cobra"
)

// NewWhoamiCommand creates the 'aerostack whoami' command
func NewWhoamiCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "whoami",
		Short: "Show current Aerostack login and linked project",
		Long:  `Display the currently logged-in Aerostack account and linked project.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return whoami()
		},
	}

	return cmd
}

func whoami() error {
	cred, err := credentials.Load()
	if err != nil {
		return fmt.Errorf("load credentials: %w", err)
	}

	if cred == nil || cred.APIKey == "" {
		fmt.Println("Not logged in. Run 'aerostack login'")
		return nil
	}

	resp, err := api.Validate(cred.APIKey)
	if err != nil {
		fmt.Printf("Logged in but validation failed: %v\n", err)
		fmt.Println("Run 'aerostack login' to re-authenticate.")
		return nil
	}

	fmt.Println("Aerostack")
	if resp.KeyType == "account" {
		fmt.Println("  Key type: account (full access)")
		if resp.Email != "" {
			fmt.Println("  Email:", resp.Email)
		}
		fmt.Println("  Deploy: uses name from aerostack.toml, no link required")
	} else {
		projLink, _ := link.Load()
		linked := "no"
		if projLink != nil && projLink.ProjectID != "" {
			if projLink.ProjectID == resp.ProjectID {
				linked = "yes (" + resp.ProjectID + ")"
			} else {
				linked = "different project (" + projLink.ProjectID + ")"
			}
		}
		fmt.Println("  Key type: project")
		fmt.Println("  Project:", resp.ProjectName, "("+resp.Slug+")")
		fmt.Println("  ID:", resp.ProjectID)
		fmt.Println("  URL:", resp.URL)
		fmt.Println("  Linked (this directory):", linked)
	}
	return nil
}
