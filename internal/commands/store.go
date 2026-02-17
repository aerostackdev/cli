package commands

import (
	"os"

	"github.com/aerostackdev/cli/internal/agent"
	"github.com/aerostackdev/cli/internal/modules/store"
	"github.com/aerostackdev/cli/internal/pkg"
	"github.com/spf13/cobra"
)

func NewStoreCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "store",
		Short: "Manage data store resources",
	}

	storeCmdInit := func() (*agent.Agent, error) {
		cwd, _ := os.Getwd()
		pkgStore, err := pkg.NewStore(cwd)
		if err != nil {
			return nil, err
		}
		return agent.NewAgent(pkgStore, false)
	}

	schemaCmd := &cobra.Command{
		Use:   "schema [intent]",
		Short: "Generate schema changes from natural language",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ag, err := storeCmdInit()
			if err != nil {
				return err
			}
			ai := store.NewSchemaAI(ag)
			return ai.Generate(cmd.Context(), args[0])
		},
	}

	seedCmd := &cobra.Command{
		Use:   "seed",
		Short: "Seed the database with data",
		RunE: func(cmd *cobra.Command, args []string) error {
			smart, _ := cmd.Flags().GetBool("smart")
			ag, err := storeCmdInit()
			if err != nil {
				return err
			}
			ai := store.NewSeederAI(ag)
			return ai.Seed(cmd.Context(), smart)
		},
	}
	seedCmd.Flags().Bool("smart", false, "Use AI to generate realistic seed data based on schema")

	cmd.AddCommand(schemaCmd)
	cmd.AddCommand(seedCmd)
	return cmd
}
