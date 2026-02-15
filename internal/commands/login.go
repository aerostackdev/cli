package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/aerostackdev/cli/internal/api"
	"github.com/aerostackdev/cli/internal/credentials"
	"github.com/spf13/cobra"
)

// NewLoginCommand creates the 'aerostack login' command
func NewLoginCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with Aerostack",
		Long: `Authenticate with your Aerostack account using a project API key.

Get your API key from the Aerostack dashboard: create a project, then add an API key
in Project Settings. Use that key here to enable deploy to Aerostack.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return login()
		},
	}

	return cmd
}

func login() error {
	apiKey := os.Getenv("AEROSTACK_API_KEY")
	if apiKey == "" {
		fmt.Print("Enter your Aerostack API key (ak_...): ")
		reader := bufio.NewReader(os.Stdin)
		line, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("read input: %w", err)
		}
		apiKey = strings.TrimSpace(line)
		if apiKey == "" {
			return fmt.Errorf("API key is required")
		}
	}

	fmt.Println("🔐 Validating API key...")
	resp, err := api.Validate(apiKey)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if err := credentials.Save(apiKey); err != nil {
		return fmt.Errorf("save credentials: %w", err)
	}

	if resp.KeyType == "account" {
		fmt.Printf("✅ Logged in! Account key (full access)\n")
		if resp.Email != "" {
			fmt.Printf("   Email: %s\n", resp.Email)
		}
		fmt.Printf("   Deploy with name from aerostack.toml — no link required\n")
	} else {
		fmt.Printf("✅ Logged in! Project: %s (%s)\n", resp.ProjectName, resp.Slug)
		fmt.Printf("   URL: %s\n", resp.URL)
	}
	return nil
}
