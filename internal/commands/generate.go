package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aerostackdev/cli/internal/api"
	"github.com/aerostackdev/cli/internal/devserver"
	"github.com/spf13/cobra"
)

// NewGenerateCommand creates the 'aerostack generate' command
func NewGenerateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate code and types",
		Long:  `Generate TypeScript types, database clients, and other project artifacts.`,
	}

	cmd.AddCommand(newGenerateTypesCommand())

	return cmd
}

func newGenerateTypesCommand() *cobra.Command {
	var outputPath string

	cmd := &cobra.Command{
		Use:   "types",
		Short: "Generate TypeScript types from database schema",
		Long: `Introspects all connected databases (D1 and Postgres) and generates 
TypeScript interfaces and a type-safe database client.

Example:
  aerostack generate types --output src/db/types.ts`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return generateTypes(outputPath)
		},
	}

	cmd.Flags().StringVarP(&outputPath, "output", "o", "shared/types.ts", "File path for generated types")

	return cmd
}

func generateTypes(outputPath string) error {
	fmt.Println("📊 Starting deep introspection...")

	// 1. Parse aerostack.toml
	cfg, err := devserver.ParseAerostackToml("aerostack.toml")
	if err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	projectRoot, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get project root: %w", err)
	}

	// 2. Fetch Project Metadata (Collections, Hooks, Queues, etc.)
	apiKey := os.Getenv("AEROSTACK_API_KEY")
	var metadata *api.ProjectMetadata
	if apiKey != "" {
		fmt.Println("🛰️  Fetching project metadata from Aerostack API...")
		meta, err := api.GetProjectMetadata(apiKey, cfg.ProjectSlug)
		if err != nil {
			fmt.Printf("⚠️  Metadata fetch warning: %v\n", err)
		} else {
			metadata = meta
			fmt.Printf("✅ Metadata fetched: %d collections, %d hooks\n", len(meta.Collections), len(meta.Hooks))
		}
	} else {
		fmt.Println("ℹ️  AEROSTACK_API_KEY not set. Skipping deep resource introspection (collections, hooks, etc.)")
	}

	// 3. Ensure wrangler.toml exists for D1 introspection (wrangler needs it)
	devserver.EnsureDefaultD1(cfg)

	wranglerPath := filepath.Join(projectRoot, "wrangler.toml")
	if _, err := os.Stat(wranglerPath); os.IsNotExist(err) {
		if err := devserver.GenerateWranglerToml(cfg, wranglerPath); err != nil {
			return fmt.Errorf("failed to generate wrangler.toml for D1 introspection: %w", err)
		}
		fmt.Println("📄 Generated wrangler.toml for D1 introspection")
	}

	var allSchemas []devserver.TableSchema

	// 4. Introspect D1 (if any)
	for _, d1 := range cfg.D1Databases {
		fmt.Printf("🔍 Introspecting D1 %s (%s)...\n", d1.Binding, d1.DatabaseName)
		d1Schemas, err := devserver.IntrospectD1Local(d1.DatabaseName, projectRoot, d1.Binding)
		if err != nil {
			fmt.Printf("⚠️  D1 introspection warning: %v\n", err)
		} else {
			allSchemas = append(allSchemas, d1Schemas...)
		}
	}

	// 5. Introspect Postgres (if any)
	for _, pg := range cfg.PostgresDatabases {
		if strings.Contains(pg.ConnectionString, "$") {
			return fmt.Errorf("Postgres binding %q: connection string has unresolved env vars. Set the required env var (e.g. in .env) and try again", pg.Binding)
		}
		fmt.Printf("🔍 Introspecting Postgres (%s)...\n", pg.Binding)
		pgSchemas, err := devserver.IntrospectPostgres(pg.ConnectionString, pg.Binding)
		if err != nil {
			fmt.Printf("⚠️  Postgres introspection warning: %v\n", err)
		} else {
			allSchemas = append(allSchemas, pgSchemas...)
		}
	}

	if len(allSchemas) == 0 && metadata == nil {
		return fmt.Errorf("no resources found to generate types from (no databases and no metadata)")
	}

	// 6. Generate TypeScript
	tsCode := devserver.GenerateTypeScript(allSchemas, metadata)

	// 7. Ensure directory exists and write file
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory for types: %w", err)
	}

	if err := os.WriteFile(outputPath, []byte(tsCode), 0644); err != nil {
		return fmt.Errorf("failed to write types file: %w", err)
	}

	totalTables := len(allSchemas)
	fmt.Printf("✨ Generated IntelliSense types (%d tables) → %s\n", totalTables, outputPath)

	return nil
}
