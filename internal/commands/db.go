package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/aerostackdev/cli/internal/devserver"
	"github.com/aerostackdev/cli/internal/neon"
	"github.com/spf13/cobra"
)

// NewDBCommand creates the 'aerostack db' command
func NewDBCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "db",
		Short: "Manage databases",
		Long: `Manage D1 and external Postgres databases.

Commands:
  aerostack db neon create <name>  Create a new Neon Postgres database
  aerostack db migrate new <name>  Create a new migration file
  aerostack db migrate apply       Apply pending migrations`,
	}

	// Add neon subcommand
	cmd.AddCommand(newNeonCommand())
	// Add migrate subcommand
	cmd.AddCommand(newDBMigrateCommand())
	// Add pull subcommand (alias for generate types)
	cmd.AddCommand(newDBPullCommand())

	return cmd
}

func newDBPullCommand() *cobra.Command {
	var outputPath string
	cmd := &cobra.Command{
		Use:   "pull",
		Short: "Pull schema and generate TypeScript types (alias for generate types)",
		Long:  `Introspects all connected databases (D1 and Postgres) and generates TypeScript interfaces. Same as 'aerostack generate types'.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return generateTypes(outputPath)
		},
	}
	cmd.Flags().StringVarP(&outputPath, "output", "o", "shared/types.ts", "Output path for generated types")
	return cmd
}

func newDBMigrateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Manage database migrations",
	}

	cmd.AddCommand(newMigrateCreateCommand())
	cmd.AddCommand(newMigrateApplyCommand())

	return cmd
}

func newMigrateCreateCommand() *cobra.Command {
	var postgres bool
	cmd := &cobra.Command{
		Use:   "new <name>",
		Short: "Create a new migration file",
		Long:  "Creates migrations/ for D1 (default) or migrations_postgres/ with --postgres",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return createMigration(args[0], postgres)
		},
	}
	cmd.Flags().BoolVar(&postgres, "postgres", false, "Create migration in migrations_postgres/ for Postgres")
	return cmd
}

func newMigrateApplyCommand() *cobra.Command {
	var remote string
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply pending migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			return applyMigrations(remote)
		},
	}
	cmd.Flags().StringVar(&remote, "remote", "", "Apply to remote environment (staging/production)")
	return cmd
}

func createMigration(name string, postgres bool) error {
	dir := "migrations"
	if postgres {
		dir = "migrations_postgres"
	}
	timestamp := time.Now().Format("20060102150405") // YYYYMMDDHHmmss
	filename := fmt.Sprintf("%s/%s_%s.sql", dir, timestamp, name)

	// Ensure migrations directory exists
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create migrations directory: %w", err)
	}

	// Create the migration file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create migration file %s: %w", filename, err)
	}
	defer file.Close()

	fmt.Printf("✅ Created migration file: %s\n", filename)
	return nil
}

func applyMigrations(remote string) error {
	env := "local"
	if remote != "" {
		env = remote
	}
	fmt.Printf("Applying migrations to %s environment...\n", env)

	// 1. Parse aerostack.toml
	cfg, err := devserver.ParseAerostackToml("aerostack.toml")
	if err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	projectRoot, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get project root: %w", err)
	}

	// 2. Ensure D1 databases, but ONLY if no Postgres is configured
	// (EnsureDefaultD1 now handles this correctly internally)
	devserver.EnsureDefaultD1(cfg)

	// 3. Apply D1 migrations (if migrations/ exists and D1 is configured)
	useRemote := remote != ""
	hasD1Migrations := false
	if _, err := os.Stat("migrations"); err == nil {
		hasD1Migrations = true
	}
	hasMigrated := false
	if hasD1Migrations && len(cfg.D1Databases) > 0 {
		// Need wrangler.toml for D1 migrations
		wranglerPath := filepath.Join(projectRoot, "wrangler.toml")
		if err := devserver.GenerateWranglerToml(cfg, wranglerPath); err != nil {
			return fmt.Errorf("failed to generate wrangler.toml: %w", err)
		}
		for _, db := range cfg.D1Databases {
			fmt.Printf("📦 Applying migrations to D1 %s (%s)...\n", db.Binding, db.DatabaseName)
			args := []string{"-y", "wrangler@latest", "d1", "migrations", "apply", db.DatabaseName}
			if useRemote {
				args = append(args, "--remote")
			} else {
				args = append(args, "--local")
			}
			cmd := exec.Command("npx", args...)
			cmd.Dir = projectRoot
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("D1 migrations failed for %s: %w", db.DatabaseName, err)
			}
			hasMigrated = true
		}
	}

	// 4. Apply Postgres migrations (migrations_postgres/*.sql)
	// For remote: look for DATABASE_URL or per-binding env var as connection string override
	hasPostgresMigrations := false
	if _, err := os.Stat("migrations_postgres"); err == nil {
		hasPostgresMigrations = true
	}
	if hasPostgresMigrations && len(cfg.PostgresDatabases) > 0 {
		for _, pg := range cfg.PostgresDatabases {
			connStr := pg.ConnectionString

			// For remote migrations, allow overriding connection string from env
			if useRemote {
				// Try env var override: e.g. PRODUCTION_PG_CONN or DATABASE_URL
				envVarName := strings.ToUpper(pg.Binding) + "_CONN"
				if override := os.Getenv(envVarName); override != "" {
					connStr = override
				} else if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
					connStr = dbURL
				}
			}

			if strings.Contains(connStr, "$") {
				fmt.Printf("⚠️  Skipping Postgres %s: connection string has unresolved env vars. Set DATABASE_URL or %s_CONN in your environment.\n", pg.Binding, strings.ToUpper(pg.Binding))
				continue
			}
			fmt.Printf("📦 Applying Postgres migrations to %s (%s env)...\n", pg.Binding, env)
			n, err := devserver.ApplyPostgresMigrations(connStr, pg.Binding)
			if err != nil {
				return fmt.Errorf("Postgres migrations failed for %s: %w", pg.Binding, err)
			}
			if n > 0 {
				fmt.Printf("   ✓ Applied %d migration(s)\n", n)
			} else {
				fmt.Printf("   ✓ No new migrations to apply\n")
			}
			hasMigrated = true
		}
	}

	// 5. Summary
	if !hasMigrated && !hasD1Migrations && !hasPostgresMigrations {
		fmt.Println("No migrations directory found.")
		fmt.Println("  For D1:       aerostack db migrate new <name>")
		fmt.Println("  For Postgres: aerostack db migrate new <name> --postgres")
	} else if !hasMigrated {
		fmt.Println("No applicable migrations found for the current configuration.")
	} else {
		fmt.Printf("✅ Migrations applied to %s\n", env)
	}
	return nil
}

func newNeonCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "neon",
		Short: "Manage Neon Postgres databases",
	}

	// neon create command
	var region string
	var addToConfig bool

	createCmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new Neon Postgres database",
		Long: `Create a new managed Postgres database on Neon.

Requires NEON_API_KEY environment variable. Get your API key from:
https://console.neon.tech/app/settings/api-keys

Example:
  export NEON_API_KEY=your-api-key
  aerostack db neon create my-database --region us-west-2 --add-to-config`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return createNeonDatabase(args[0], region, addToConfig)
		},
	}

	createCmd.Flags().StringVar(&region, "region", "aws-us-west-2", "Neon region (aws-us-west-2, aws-us-east-1, aws-eu-central-1, etc.)")
	createCmd.Flags().BoolVar(&addToConfig, "add-to-config", true, "Automatically add to aerostack.toml")

	cmd.AddCommand(createCmd)
	return cmd
}

func createNeonDatabase(name, region string, addToConfig bool) error {
	fmt.Println("🚀 Creating Neon Postgres database...")

	// 1. Get Neon API key
	apiKey, err := neon.GetAPIKeyFromEnv()
	if err != nil {
		return err
	}

	// 2. Create Neon project
	client := neon.NewClient(apiKey)
	result, err := client.CreateProject(name, region)
	if err != nil {
		return fmt.Errorf("failed to create Neon project: %w", err)
	}

	// 3. Get connection string (API may return URI directly or in nested structure)
	connStr := result.GetConnectionString()
	if connStr == "" {
		return fmt.Errorf("Neon API did not return a connection URI. Check your API key and try again")
	}

	fmt.Println("✅ Neon database created successfully!")
	fmt.Printf("   Project ID: %s\n", result.Project.ID)
	fmt.Printf("   Region: %s\n", result.Project.RegionID)
	fmt.Println()

	// 4. Show connection info
	envVarName := strings.ToUpper(strings.ReplaceAll(name, "-", "_")) + "_DATABASE_URL"
	fmt.Printf("📋 Connection String (save to .env):\n")
	fmt.Printf("   %s=\"%s\"\n", envVarName, connStr)
	fmt.Println()

	// 5. Add to aerostack.toml if requested
	if addToConfig {
		if err := addPostgresToConfig(name, envVarName); err != nil {
			fmt.Printf("⚠️  Failed to update aerostack.toml: %v\n", err)
			fmt.Println("   Please add manually:")
			showManualConfigInstructions(name, envVarName)
		} else {
			fmt.Println("✅ Added to aerostack.toml")
			showManualConfigInstructions(name, envVarName)
		}
	} else {
		showManualConfigInstructions(name, envVarName)
	}

	return nil
}

func addPostgresToConfig(name, envVarName string) error {
	if name == "" {
		return fmt.Errorf("database name cannot be empty")
	}
	// Check if aerostack.toml exists
	if _, err := os.Stat("aerostack.toml"); os.IsNotExist(err) {
		return fmt.Errorf("aerostack.toml not found in current directory")
	}

	// Read existing config
	data, err := os.ReadFile("aerostack.toml")
	if err != nil {
		return err
	}

	// Append Postgres configuration
	binding := strings.ToUpper(name[:1]) + strings.ToLower(name[1:])
	if len(binding) > 10 {
		binding = strings.ToUpper(name[:2]) + "DB"
	}

	postgresBlock := fmt.Sprintf(`

# Neon Postgres Database
[[postgres_databases]]
binding = "%s"
connection_string = "$%s"
pool_size = 10
`, binding, envVarName)

	// Append to file
	newData := append(data, []byte(postgresBlock)...)
	if err := os.WriteFile("aerostack.toml", newData, 0644); err != nil {
		return err
	}

	return nil
}

func showManualConfigInstructions(name, envVarName string) {
	binding := strings.ToUpper(name[:1]) + strings.ToLower(name[1:])
	if len(binding) > 10 {
		binding = strings.ToUpper(name[:2]) + "DB"
	}

	fmt.Println("\n📝 Next steps:")
	fmt.Printf("   1. Add %s to your .env file\n", envVarName)
	fmt.Println("   2. Use in your code:")
	fmt.Printf("      const db = env.%s // TypeScript\n", binding)
	fmt.Println("   3. Run 'aerostack dev' to start local development")
}
