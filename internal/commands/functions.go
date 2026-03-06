package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/aerostackdev/cli/internal/api"
	"github.com/aerostackdev/cli/internal/credentials"
	"github.com/aerostackdev/cli/internal/pkg"
	"github.com/spf13/cobra"
)

// slugifyName normalizes a name to a slug (lowercase, hyphens). Must match API slugify for upsert.
func slugifyName(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	s = regexp.MustCompile(`[^\w\s-]`).ReplaceAllString(s, "")
	s = regexp.MustCompile(`[\s_]+`).ReplaceAllString(s, "-")
	s = regexp.MustCompile(`-+`).ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

// NewFunctionsCommand creates the 'aerostack functions' root command
func NewFunctionsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "functions",
		Short: "Manage community functions",
		Long: `Create, push, pull, and publish community functions in the Aerostack registry.
Functions are reusable units of logic that can be shared across the community.`,
	}

	cmd.AddCommand(NewFunctionsPushCommand())
	cmd.AddCommand(NewFunctionsPullCommand())
	cmd.AddCommand(NewFunctionsPublishCommand())
	cmd.AddCommand(NewFunctionsInstallCommand())
	return cmd
}

// NewFunctionsInstallCommand creates 'aerostack functions install [username/slug]'
func NewFunctionsInstallCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "install [username/slug]",
		Short: "Install a community function into your project",
		Long: `Install an OFS function from the Aerostack registry into src/modules/<slug>/.

Files are copied directly into your project (not added as node_modules).
For Hono/Cloudflare Workers projects the route is registered automatically.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref := args[0]
			var username, slug string

			if strings.Contains(ref, "/") {
				parts := strings.Split(ref, "/")
				username = strings.TrimPrefix(parts[0], "@")
				slug = parts[1]
			} else {
				fmt.Printf("🔍 Searching for function '%s'...\n", ref)
				slug = ref
			}

			// 1. Fetch the full OFS install manifest from the registry
			var manifest *api.InstallManifest
			var err error
			if username != "" {
				manifest, err = api.CommunityGetInstallManifest(username, slug)
			} else {
				manifest, err = api.CommunityGetInstallManifestBySlug(slug)
			}
			if err != nil {
				return err
			}

			fmt.Printf("📥 Installing %s/%s v%s...\n", manifest.Author, manifest.Slug, manifest.Version)

			// 2. Write each file to src/modules/<slug>/
			moduleDir := filepath.Join("src", "modules", manifest.Slug)
			if err := os.MkdirAll(moduleDir, 0755); err != nil {
				return fmt.Errorf("failed to create module directory: %w", err)
			}

			for _, f := range manifest.Files {
				destPath := filepath.Join(moduleDir, filepath.FromSlash(f.Path))
				// Ensure subdirectory exists
				if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
					return fmt.Errorf("failed to create directory for %s: %w", f.Path, err)
				}
				if err := os.WriteFile(destPath, []byte(f.Content), 0644); err != nil {
					return fmt.Errorf("failed to write %s: %w", f.Path, err)
				}
				fmt.Printf("   + %s\n", filepath.Join("src/modules", manifest.Slug, f.Path))
			}

			// 3. Write aerostack-manifest.json (install metadata for tooling)
			manifestJSON, _ := json.MarshalIndent(map[string]interface{}{
				"name":            manifest.Name,
				"slug":            manifest.Slug,
				"version":         manifest.Version,
				"author":          manifest.Author,
				"routeExport":     manifest.RouteExport,
				"routePath":       manifest.RoutePath,
				"npmDependencies": manifest.NpmDependencies,
				"envVars":         manifest.EnvVars,
				"drizzleSchema":   manifest.DrizzleSchema,
			}, "", "  ")
			_ = os.WriteFile(filepath.Join(moduleDir, "aerostack-manifest.json"), manifestJSON, 0644)

			// 4. Register in aerostack.toml
			indexPath := filepath.Join(moduleDir, "index.ts")
			if err := pkg.AppendServiceToAerostackToml(manifest.Slug, indexPath); err != nil {
				fmt.Printf("⚠️  Warning: Failed to register in aerostack.toml: %v\n", err)
			}

			// 5. Patch src/index.ts with import + route registration
			if err := patchIndexTs(manifest); err != nil {
				fmt.Printf("⚠️  Could not auto-wire into src/index.ts: %v\n", err)
				fmt.Printf("   Add manually:\n")
				fmt.Printf("   import { %s } from './modules/%s';\n", manifest.RouteExport, manifest.Slug)
				fmt.Printf("   app.route('%s', %s);\n", manifest.RoutePath, manifest.RouteExport)
			}

			// 6. Print npm install reminder if dependencies are needed
			if len(manifest.NpmDependencies) > 0 {
				fmt.Printf("\n📦 Run: npm install %s\n", strings.Join(manifest.NpmDependencies, " "))
			}

			// 7. Print env var reminder
			if len(manifest.EnvVars) > 0 {
				fmt.Printf("🔑 Add to .dev.vars: %s\n", strings.Join(manifest.EnvVars, ", "))
			}

			// 8. DB schema notice
			if manifest.DrizzleSchema {
				fmt.Printf("🗄️  This function includes a DB schema. Run 'aerostack db migrate' to apply it.\n")
			}

			fmt.Printf("\n✅ Installed %s/%s to %s\n", manifest.Author, manifest.Slug, moduleDir)
			return nil
		},
	}
}

// patchIndexTs inserts the import and route registration into the consumer's src/index.ts.
// It looks for // aerostack:imports and // aerostack:routes comment markers.
func patchIndexTs(manifest *api.InstallManifest) error {
	indexPath := filepath.Join("src", "index.ts")
	content, err := os.ReadFile(indexPath)
	if err != nil {
		return fmt.Errorf("src/index.ts not found")
	}

	src := string(content)
	importLine := fmt.Sprintf("import { %s } from './modules/%s';", manifest.RouteExport, manifest.Slug)
	routeLine := fmt.Sprintf("app.route('%s', %s);", manifest.RoutePath, manifest.RouteExport)

	// Avoid duplicate inserts
	if strings.Contains(src, importLine) {
		return nil
	}

	importMarker := "// aerostack:imports"
	routeMarker := "// aerostack:routes"

	src = strings.Replace(src, importMarker, importMarker+"\n"+importLine, 1)
	src = strings.Replace(src, routeMarker, routeMarker+"\n"+routeLine, 1)

	return os.WriteFile(indexPath, []byte(src), 0644)
}

// NewFunctionsPushCommand creates 'aerostack functions push'
func NewFunctionsPushCommand() *cobra.Command {
	var name string
	var category string
	var description string
	var tags []string

	cmd := &cobra.Command{
		Use:   "push [file]",
		Short: "Push a local function to the Aerostack community registry (as draft)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := "src/index.ts"
			if len(args) > 0 {
				filePath = args[0]
			}

			// 1. Read file
			code, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("failed to read %s: %w", filePath, err)
			}

			// 2. Load API Key
			apiKey := credentials.GetAPIKey()
			if apiKey == "" {
				return fmt.Errorf("not logged in. Run 'aerostack login' first")
			}

			// 3. Prepare function metadata
			if name == "" {
				// Try to get from folder name or file name
				name = filepath.Base(filepath.Dir(filePath))
				if name == "." || name == "src" {
					cwd, _ := os.Getwd()
					name = filepath.Base(cwd)
				}
			}

			// 4. Push to API
			fn := api.CommunityFunction{
				Name:        name,
				Description: description,
				Category:    category,
				Code:        string(code),
				Tags:        tags,
				Language:    "typescript",
				Runtime:     "cloudflare-worker",
				Files:       make(map[string]string),
			}

			// Read aerostack.json configuration. Check file directory then CWD.
			configPaths := []string{
				filepath.Join(filepath.Dir(filePath), "aerostack.json"),
				"aerostack.json",
			}
			var configContent []byte
			var configPath string
			for _, cp := range configPaths {
				if content, err := os.ReadFile(cp); err == nil {
					configContent = content
					configPath = cp
					break
				}
			}

			if configContent != nil {
				var configData map[string]interface{}
				if err := json.Unmarshal(configContent, &configData); err != nil {
					return fmt.Errorf("invalid aerostack.json: %w", err)
				}
				fn.ConfigSchema = configData

				// Override metadata from config if present
				if confName, ok := configData["name"].(string); ok && confName != "" {
					fn.Name = confName
				}
				if confDesc, ok := configData["description"].(string); ok && confDesc != "" {
					fn.Description = confDesc
				}
				if confCat, ok := configData["category"].(string); ok && confCat != "" {
					fn.Category = confCat
				} else if confType, ok := configData["type"].(string); ok && confType != "" {
					fn.Category = confType // Legacy/Fallback
				}
				if confVersion, ok := configData["version"].(string); ok && confVersion != "" {
					fn.Version = confVersion
				}

				// OFS v1 Multi-file support: Collect files if schema matches
				if schema, ok := configData["$schema"].(string); ok && strings.Contains(schema, "ofs-v1.json") {
					fn.Files["aerostack.json"] = string(configContent)
					// Collect standard OFS files if they exist
					baseDir := filepath.Dir(configPath)
					ofsFiles := []string{
						"src/core.ts",
						"src/adapter.ts",
						"src/node-adapter.ts",
						"src/index.ts",
						"src/schema.ts",
						"README.md",
						"package.json",
					}
					for _, relPath := range ofsFiles {
						fullPath := filepath.Join(baseDir, relPath)
						if data, err := os.ReadFile(fullPath); err == nil {
							fn.Files[relPath] = string(data)
						}
					}
					// Ensure 'Code' is set to core.ts if it exists
					if coreCode, ok := fn.Files["src/core.ts"]; ok {
						fn.Code = coreCode
					}
				}

				// Update local name variable for logging
				name = fn.Name
			}
			// Slug for upsert: API matches by (developer, slug); derive from name so push updates same function
			fn.Slug = slugifyName(fn.Name)
			if fn.Version == "" {
				fn.Version = "1.0.0"
			}

			// Try to read README.md in the same directory (fallback if not already in fn.Files)
			if _, ok := fn.Files["README.md"]; !ok {
				readmePath := filepath.Join(filepath.Dir(filePath), "README.md")
				if readmeContent, err := os.ReadFile(readmePath); err == nil {
					fn.Readme = string(readmeContent)
					fn.Files["README.md"] = string(readmeContent)
				}
			} else {
				fn.Readme = fn.Files["README.md"]
			}

			fmt.Printf("🚀 Pushing function '%s' to Aerostack...\n", fn.Name)
			resp, err := api.CommunityPush(apiKey, fn)
			if err != nil {
				return err
			}

			fmt.Printf("✅ Pushed successfully!\n")
			fmt.Printf("   Slug: %s\n", resp.Slug)
			fmt.Printf("   Status: %s\n", resp.Status)
			fmt.Printf("   Admin URL: https://admin.aerocall.ai/functions/edit/%s\n", resp.ID)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Name of the function")
	cmd.Flags().StringVar(&category, "category", "utility", "Category (auth, payments, ai, etc.)")
	cmd.Flags().StringVar(&description, "description", "A reusable Aerostack function", "Short description")
	cmd.Flags().StringSliceVar(&tags, "tags", []string{}, "Comma-separated tags")

	return cmd
}

// NewFunctionsPullCommand creates 'aerostack functions pull'
func NewFunctionsPullCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "pull [username/slug]",
		Short: "Pull a function's source code from the community registry",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			parts := strings.Split(args[0], "/")
			if len(parts) != 2 {
				return fmt.Errorf("invalid format. Use: username/slug (e.g. alice/stripe-checkout)")
			}
			username, slug := parts[0], parts[1]

			fmt.Printf("📥 Pulling %s/%s...\n", username, slug)
			fn, err := api.CommunityPull(username, slug)
			if err != nil {
				return err
			}

			// 1. Create a directory for the function
			if err := os.MkdirAll(slug, 0755); err != nil {
				return err
			}

			// 2. Write source
			sourcePath := filepath.Join(slug, "index.ts")
			if err := os.WriteFile(sourcePath, []byte(fn.Code), 0644); err != nil {
				return err
			}

			// 3. Write readme if exists
			if fn.Readme != "" {
				readmePath := filepath.Join(slug, "README.md")
				_ = os.WriteFile(readmePath, []byte(fn.Readme), 0644)
			}

			fmt.Printf("✅ Pulled to ./%s/\n", slug)
			return nil
		},
	}
}

// NewFunctionsPublishCommand creates 'aerostack functions publish'
func NewFunctionsPublishCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "publish [id-or-slug]",
		Short: "Publish a draft function to the community registry",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			// 2. Load API Key
			apiKey := credentials.GetAPIKey()
			if apiKey == "" {
				return fmt.Errorf("not logged in. Run 'aerostack login' first")
			}

			fmt.Printf("🚀 Publishing function %s...\n", id)
			if err := api.CommunityPublish(apiKey, id); err != nil {
				return err
			}

			fmt.Printf("✅ Published successfully!\n")
			return nil
		},
	}
}
