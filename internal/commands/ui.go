package commands

import (
	"fmt"
	"os"

	"github.com/aerostackdev/cli/internal/agent"
	"github.com/aerostackdev/cli/internal/modules/ui"
	"github.com/aerostackdev/cli/internal/pkg"
	"github.com/spf13/cobra"
)

func NewUICommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ui",
		Short: "Manage UI resources",
	}

	syncCmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync UI theme configuration for the AI Agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, _ := os.Getwd()
			store, err := pkg.NewStore(cwd)
			if err != nil {
				return fmt.Errorf("failed to load PKG: %w", err)
			}

			ag, err := agent.NewAgent(store, false)
			if err != nil {
				return fmt.Errorf("failed to init AI agent: %w", err)
			}

			syncer := ui.NewUISync(ag)
			return syncer.Sync(cmd.Context())
		},
	}

	cmd.AddCommand(syncCmd)
	return cmd
}
