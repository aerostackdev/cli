package commands

import (
	"fmt"
	"os"

	"github.com/aerostackdev/cli/internal/pkg"
	"github.com/spf13/cobra"
)

// NewIndexCommand creates the 'aerostack index' command
func NewIndexCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "index",
		Short: "Index the current project",
		Long:  `Scans the current project to build the Project Knowledge Graph (PKG).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}

			fmt.Println("Initializing PKG store...")
			store, err := pkg.NewStore(cwd)
			if err != nil {
				return err
			}
			defer store.Close()

			fmt.Println("Indexing project symbols...")
			indexer := pkg.NewIndexer(store)
			if err := indexer.IndexProject(cwd); err != nil {
				return err
			}

			fmt.Println("Mapping relationships...")
			mapper := pkg.NewMapper(store)
			if err := mapper.MapRelationships(); err != nil {
				return err
			}

			fmt.Println("Project indexed successfully.")
			return nil
		},
	}
	return cmd
}
