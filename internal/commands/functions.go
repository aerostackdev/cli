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
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]
			var username, slug string

			if strings.Contains(path, "/") {
				parts := strings.Split(path, "/")
				username = strings.TrimPrefix(parts[0], "@")
				slug = parts[1]
			} else {
				// Search by slug only (picking best match)
				fmt.Printf("🔍 Searching for function '%s'...\n", path)
				slug = path
			}

			// 1. Pull manifest/code from API
			var fn *api.CommunityFunction
			var err error
			if username != "" {
				fn, err = api.CommunityPull(username, slug)
			} else {
				// Use the convenience endpoint (already implemented in API as /install/:slug)
				// I'll add a new API client method for this.
				fn, err = api.CommunityPullNoAuth(slug)
			}
			if err != nil {
				return err
			}

			fmt.Printf("📥 Installing %s/%s...\n", fn.AuthorUsername, fn.Slug)

			// 2. Create directory in services/
			serviceDir := filepath.Join("services", fn.Slug)
			if err := os.MkdirAll(serviceDir, 0755); err != nil {
				return err
			}

			// 3. Write source
			indexPath := filepath.Join(serviceDir, "index.ts")
			if err := os.WriteFile(indexPath, []byte(fn.Code), 0644); err != nil {
				return err
			}

			// 4. Write README
			if fn.Readme != "" {
				readmePath := filepath.Join(serviceDir, "README.md")
				_ = os.WriteFile(readmePath, []byte(fn.Readme), 0644)
			}

			// 5. Register in aerostack.toml
			if err := pkg.AppendServiceToAerostackToml(fn.Slug, indexPath); err != nil {
				fmt.Printf("⚠️  Warning: Failed to register in aerostack.toml: %v\n", err)
			}

			fmt.Printf("✅ Installed to %s\n", serviceDir)
			return nil
		},
	}
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
			}

			// Read aerostack.json configuration. Check file directory then CWD.
			configPaths := []string{
				filepath.Join(filepath.Dir(filePath), "aerostack.json"),
				"aerostack.json",
			}
			var configContent []byte
			for _, cp := range configPaths {
				if content, err := os.ReadFile(cp); err == nil {
					configContent = content
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
				if confType, ok := configData["type"].(string); ok && confType != "" {
					fn.Category = confType // Maps schema 'type' to DB 'category'
				}
				if confVersion, ok := configData["version"].(string); ok && confVersion != "" {
					fn.Version = confVersion
				}

				// Update local name variable for logging
				name = fn.Name
			}
			// Slug for upsert: API matches by (developer, slug); derive from name so push updates same function
			fn.Slug = slugifyName(fn.Name)
			if fn.Version == "" {
				fn.Version = "1.0.0"
			}

			// Try to read README.md in the same directory
			readmePath := filepath.Join(filepath.Dir(filePath), "README.md")
			if readmeContent, err := os.ReadFile(readmePath); err == nil {
				fn.Readme = string(readmeContent)
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
