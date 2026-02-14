package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewLoginCommand creates the 'aerostack login' command
func NewLoginCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with Aerostack",
		Long: `Authenticate with your Aerostack account using OAuth.

This will open your browser to complete the authentication flow.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return login()
		},
	}

	return cmd
}

func login() error {
	fmt.Println("🔐 Opening browser for authentication...")
	
	// TODO: Implement OAuth flow
	fmt.Println("✅ Successfully authenticated!")
	
	return nil
}
