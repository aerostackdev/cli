package commands

import (
	"fmt"
	"os"

	"github.com/aerostackdev/cli/internal/agent"
	"github.com/aerostackdev/cli/internal/pkg"
	"github.com/spf13/cobra"
)

func NewAICommand() *cobra.Command {
	var debug bool

	cmd := &cobra.Command{
		Use:   "ai [prompt]",
		Short: "Ask the AI agent explicitly",
		Long:  `Send a prompt to the Aerostack AI agent. The agent has access to your project context.`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			prompt := args[0]
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}

			// 1. Initialize PKG
			store, err := pkg.NewStore(cwd)
			if err != nil {
				return fmt.Errorf("failed to open PKG: %w (try running 'aerostack index' first)", err)
			}
			defer store.Close()

			// 2. Initialize Agent
			ag, err := agent.NewAgent(store, debug)
			if err != nil {
				return fmt.Errorf("failed to initialize agent: %w", err)
			}

			// 3. Resolve Intent
			if err := ag.Resolve(cmd.Context(), prompt); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&debug, "debug", false, "Enable debug logging")

	return cmd
}
