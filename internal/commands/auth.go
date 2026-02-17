package commands

import (
	"fmt"
	"os"

	"github.com/aerostackdev/cli/internal/agent"
	"github.com/aerostackdev/cli/internal/modules/auth"
	"github.com/aerostackdev/cli/internal/pkg"
	"github.com/spf13/cobra"
)

func NewAuthCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication resources",
	}

	doctorCmd := &cobra.Command{
		Use:   "doctor",
		Short: "Diagnose authentication issues",
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

			doc := auth.NewDoctor(ag)
			return doc.Run(cmd.Context())
		},
	}

	cmd.AddCommand(doctorCmd)
	return cmd
}
