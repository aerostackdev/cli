package commands

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"encoding/json"

	"github.com/aerostackdev/cli/internal/templates"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

// NewInitCommand creates the 'aerostack init' command
func NewInitCommand() *cobra.Command {
	var template string
	var db string
	var runDev bool

	cmd := &cobra.Command{
		Use:   "init [project-name]",
		Short: "Initialize a new Aerostack project",
		Long: `Initialize a new Aerostack project with your choice of starter template.

Available templates:
  • blank               - Minimal Worker (default)
  • api                 - REST API with Hono
  • api-neon            - REST API with Hono and Neon
  • multi-func          - Multi-function sharing code
  • cron-neon           - Scheduled task with Neon
  • webhook-neon        - Webhook processor with Neon
  • ws-voice-agent      - WebSocket voice/chat agent
  • ws-multiplayer-game - WebSocket multiplayer game sample
  • ws-chat             - WebSocket group chat (extend to 1:1 later)

Examples:
  aerostack init my-app --template=api --db=neon
  aerostack init my-app --dev    # init then start dev server (no cd needed)`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var projectName string
			if len(args) > 0 {
				projectName = args[0]
			}

			// Only run the form for missing inputs
			// We need to construct the form dynamically based on what's missing
			// But huh form structure is static-ish. Let's rebuild the form based on needs.

			var questions []huh.Field

			if len(args) == 0 {
				questions = append(questions, huh.NewInput().
					Title("Project Name").
					Description("What is the name of your new project?").
					Value(&projectName).
					Validate(func(str string) error {
						if str == "" {
							return fmt.Errorf("project name cannot be empty")
						}
						return nil
					}))
			}

			if !cmd.Flags().Changed("template") {
				questions = append(questions, huh.NewSelect[string]().
					Title("Pick a starter template").
					Description("Choose a template to bootstrap your project").
					Options(
						huh.NewOption("Blank (Minimal Worker)", "blank"),
						huh.NewOption("API (Hono)", "api"),
						huh.NewOption("API + Neon (Hono)", "api-neon"),
						huh.NewOption("Multi-Function", "multi-func"),
						huh.NewOption("Cron + Neon", "cron-neon"),
						huh.NewOption("Webhook + Neon", "webhook-neon"),
						huh.NewOption("WS Voice Agent", "ws-voice-agent"),
						huh.NewOption("WS Multiplayer Game", "ws-multiplayer-game"),
						huh.NewOption("WS Chat (group)", "ws-chat"),
					).
					Value(&template).
					WithHeight(13))
			}

			if !cmd.Flags().Changed("db") {
				questions = append(questions, huh.NewSelect[string]().
					Title("Select a database").
					Description("Which database would you like to use?").
					Options(
						huh.NewOption("Cloudflare D1 (SQLite at Edge)", "d1"),
						huh.NewOption("Neon (Serverless Postgres)", "neon"),
					).
					Value(&db))
			}

			if len(questions) > 0 {
				if err := huh.NewForm(huh.NewGroup(questions...)).Run(); err != nil {
					return err
				}
			}

			return initProject(projectName, template, db, runDev)
		},
	}

	cmd.Flags().StringVarP(&template, "template", "t", "blank", "Starter template to use")
	cmd.Flags().StringVar(&db, "db", "", "Database to use (d1, neon)")
	cmd.Flags().BoolVar(&runDev, "dev", false, "After init, run 'aerostack dev' in the new project (no need to cd)")

	return cmd
}

func initProject(name, templateName, dbName string, runDev bool) error {
	if templateName == "blank" || templateName == "" {
		templateName = "blank"
	}

	// Interactive DB selection is handled in RunE now.
	// Fallback/Default handling just in case.
	if dbName == "" {
		dbName = "d1"
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

		// 6. Update ALL dependencies to their latest versions
		fmt.Println("🔄 Updating all dependencies to latest versions...")

		// Parse package.json to get list of dependencies
		type PackageJSON struct {
			Dependencies    map[string]string `json:"dependencies"`
			DevDependencies map[string]string `json:"devDependencies"`
		}

		content, err := os.ReadFile(packageJSON)
		if err != nil {
			fmt.Printf("⚠️  Failed to read package.json for update: %v\n", err)
		} else {
			var pkg PackageJSON
			if err := json.Unmarshal(content, &pkg); err != nil {
				fmt.Printf("⚠️  Failed to parse package.json: %v\n", err)
			} else {
				// Collect all packages
				var packages []string
				for depot := range pkg.Dependencies {
					if depot == "vitest" {
						packages = append(packages, depot+"@3")
						continue
					}
					packages = append(packages, depot+"@latest")
				}
				for depot := range pkg.DevDependencies {
					if depot == "vitest" {
						packages = append(packages, depot+"@3")
						continue
					}
					packages = append(packages, depot+"@latest")
				}

				if len(packages) > 0 {
					// Install them all at once
					args := append([]string{"install"}, packages...)
					args = append(args, "--legacy-peer-deps") // Keep legacy peer deps just in case

					cmdUpdate := exec.Command("npm", args...)
					cmdUpdate.Dir = name
					cmdUpdate.Stdout = os.Stdout
					cmdUpdate.Stderr = os.Stderr
					if err := cmdUpdate.Run(); err != nil {
						fmt.Printf("⚠️  Failed to update dependencies: %v\n", err)
						// Fallback: try updating just the SDK if the bulk update fails?
						// For now, just warn.
					} else {
						fmt.Println("✨ All dependencies updated to latest!")
					}
				}
			}
		}
	}

	fmt.Println("\n✅ Project initialized successfully!")

	// Single copy-paste command so user doesn't have to cd then run dev manually
	if dbName == "neon" {
		fmt.Printf("\nNext steps (Neon): create DB, then start dev:\n")
		fmt.Printf("  cd %s && aerostack db neon create %s-db --add-to-config && aerostack dev\n", name, name)
	} else {
		fmt.Printf("\nTo start developing (one command):\n")
		fmt.Printf("  cd %s && aerostack dev\n", name)
	}

	if runDev {
		fmt.Printf("\nStarting dev server in %s...\n", name)
		exe, err := os.Executable()
		if err != nil {
			fmt.Printf("⚠️  Could not start dev: %v. Run 'cd %s && aerostack dev' manually.\n", err, name)
			return nil
		}
		devCmd := exec.Command(exe, "dev")
		devCmd.Dir = name
		devCmd.Stdin = os.Stdin
		devCmd.Stdout = os.Stdout
		devCmd.Stderr = os.Stderr
		if err := devCmd.Run(); err != nil {
			return fmt.Errorf("dev server exited: %w", err)
		}
	}

	return nil
}
