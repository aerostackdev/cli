package commands

import (
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

// NewUninstallCommand creates the 'aerostack uninstall' command
func NewUninstallCommand() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall the Aerostack CLI and remove all associated data",
		Long: `Uninstall the Aerostack CLI. 
This will:
  • Remove the $HOME/.aerostack directory (binaries and data)
  • Provide instructions for cleaning up your shell profile`,
		RunE: func(cmd *cobra.Command, args []string) error {
			confirm := force
			if !confirm {
				fmt.Println("⚠️  This will permanently delete the Aerostack CLI and all its data from your system.")
				form := huh.NewForm(
					huh.NewGroup(
						huh.NewConfirm().
							Title("Are you sure you want to uninstall?").
							Value(&confirm),
					),
				)
				if err := form.Run(); err != nil {
					return err
				}
			}

			if !confirm {
				fmt.Println("Uninstall cancelled.")
				return nil
			}

			// Remove $HOME/.aerostack
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("could not find home directory: %w", err)
			}
			aerostackDir := fmt.Sprintf("%s/.aerostack", home)

			fmt.Printf("🗑️  Removing %s...\n", aerostackDir)
			if err := os.RemoveAll(aerostackDir); err != nil {
				return fmt.Errorf("failed to remove directory: %w", err)
			}

			fmt.Println("\n✅ Aerostack CLI has been uninstalled from your home directory.")
			fmt.Println("\nNext steps (optional):")
			fmt.Println("  1. Remove the PATH entry from your shell profile (e.g., ~/.zshrc or ~/.bashrc).")
			fmt.Println("  2. The binary you just ran will remain until you close this terminal or delete it manually.")

			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "y", false, "Skip confirmation prompt")

	return cmd
}
