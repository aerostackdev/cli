package commands

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"text/template"

	"github.com/aerostackdev/cli/internal/templates"
	"github.com/spf13/cobra"
)

// NewInitCommand creates the 'aerostack init' command
func NewInitCommand() *cobra.Command {
	var template string

	cmd := &cobra.Command{
		Use:   "init [project-name]",
		Short: "Initialize a new Aerostack project",
		Long: `Initialize a new Aerostack project with your choice of starter template.

Available templates:
  • blog       - Simple blog with markdown support
  • ecommerce  - E-commerce store with Stripe integration
  • saas       - SaaS application boilerplate
  • api        - REST API with authentication
  • fullstack  - Full-stack app with React frontend

Example:
  aerostack init my-app --template=blog`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectName := args[0]
			return initProject(projectName, template)
		},
	}

	cmd.Flags().StringVarP(&template, "template", "t", "blank", "Starter template to use")

	return cmd
}

func initProject(name, templateName string) error {
	if templateName == "blank" || templateName == "" {
		templateName = "blank"
	}

	fmt.Printf("🚀 Initializing Aerostack project: %s\n", name)
	fmt.Printf("📦 Using template: %s\n", templateName)

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
	}{
		ProjectName: name,
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

	fmt.Println("\n✅ Project initialized successfully!")
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  cd %s\n", name)
	fmt.Printf("  aerostack dev\n")

	return nil
}
