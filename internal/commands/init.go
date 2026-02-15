package commands

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/aerostackdev/cli/internal/templates"
	"github.com/spf13/cobra"
)

// NewInitCommand creates the 'aerostack init' command
func NewInitCommand() *cobra.Command {
	var template string
	var db string

	cmd := &cobra.Command{
		Use:   "init [project-name]",
		Short: "Initialize a new Aerostack project",
		Long: `Initialize a new Aerostack project with your choice of starter template.

Available templates:
  • blank      - Minimal Worker (default)
  • api        - REST API with Hono
  • api-neon   - REST API with Hono and Neon
  • express    - Express.js on Workers
  • express-neon - Express.js with Neon
  • multi-func - Multi-function sharing code
  • cron-neon  - Scheduled task with Neon
  • webhook-neon - Webhook processor with Neon
  • multi-func - Multi-function sharing code

Example:
  aerostack init my-app --template=api --db=neon`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectName := args[0]
			return initProject(projectName, template, db)
		},
	}

	cmd.Flags().StringVarP(&template, "template", "t", "blank", "Starter template to use")
	cmd.Flags().StringVar(&db, "db", "", "Database to use (d1, neon)")

	return cmd
}

func initProject(name, templateName, dbName string) error {
	if templateName == "blank" || templateName == "" {
		templateName = "blank"
	}

	// Interactive DB selection if not provided
	if dbName == "" {
		fmt.Println("Choose a database for your project:")
		fmt.Println("  1. Cloudflare D1 (SQLite at the edge)")
		fmt.Println("  2. Neon PostgreSQL (Managed Postgres at the edge)")
		fmt.Print("\nSelection (1/2, default 1): ")

		var input string
		fmt.Scanln(&input)
		switch input {
		case "2":
			dbName = "neon"
		default:
			dbName = "d1"
		}
	}

	fmt.Printf("\n🚀 Initializing Aerostack project: %s\n", name)
	fmt.Printf("📦 Using template: %s\n", templateName)
	fmt.Printf("🗄️  Using database: %s\n", dbName)

	// 1. Check if template exists in embedded FS
	templatePath := filepath.Join("templates", templateName)
	if _, err := fs.ReadDir(templates.FS, templatePath); err != nil {
		return fmt.Errorf("template '%s' not found: %w", templateName, err)
	}

	// 2. Create project directory
	if err := os.MkdirAll(name, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	// 3. Define data for template placeholders
	data := struct {
		ProjectName string
		Database    string
		IsNeon      bool
		IsD1        bool
	}{
		ProjectName: name,
		Database:    dbName,
		IsNeon:      dbName == "neon",
		IsD1:        dbName == "d1",
	}

	// 4. Recursively copy and process files
	err := fs.WalkDir(templates.FS, templatePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Calculate destination path
		relPath, err := filepath.Rel(templatePath, path)
		if err != nil {
			return err
		}
		if relPath == "." {
			return nil
		}
		destPath := filepath.Join(name, relPath)

		if d.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		// Read file from embedded FS
		fileData, err := fs.ReadFile(templates.FS, path)
		if err != nil {
			return err
		}

		// Process as template
		tmpl, err := template.New(relPath).Parse(string(fileData))
		if err != nil {
			return fmt.Errorf("failed to parse template %s: %w", relPath, err)
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			return fmt.Errorf("failed to execute template %s: %w", relPath, err)
		}

		// Write to destination
		return os.WriteFile(destPath, buf.Bytes(), 0644)
	})

	if err != nil {
		return fmt.Errorf("failed to scaffold project: %w", err)
	}

	// 5. Run npm install if package.json exists
	packageJSON := filepath.Join(name, "package.json")
	if _, err := os.Stat(packageJSON); err == nil {
		fmt.Println("📦 Installing dependencies...")
		cmd := exec.Command("npm", "install", "--legacy-peer-deps")
		cmd.Dir = name
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("npm install failed: %w (run 'npm install' manually)", err)
		}
	}

	fmt.Println("\n✅ Project initialized successfully!")
	if dbName == "neon" {
		fmt.Printf("\nNext steps for Neon:\n")
		fmt.Printf("  1. cd %s\n", name)
		fmt.Printf("  2. aerostack db:neon create %s-db --add-to-config\n", name)
		fmt.Printf("  3. aerostack dev\n")
	} else {
		fmt.Printf("\nNext steps:\n")
		fmt.Printf("  1. cd %s\n", name)
		fmt.Printf("  2. aerostack dev\n")
	}

	return nil
}
