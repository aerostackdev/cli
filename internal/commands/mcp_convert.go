package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aerostackdev/cli/internal/api"
	"github.com/aerostackdev/cli/internal/credentials"
	"github.com/aerostackdev/cli/internal/modules/mcpconvert"
	"github.com/aerostackdev/cli/internal/printer"
	"github.com/spf13/cobra"
)

// NewMcpConvertCommand creates the 'aerostack mcp convert' command.
func NewMcpConvertCommand() *cobra.Command {
	var (
		packageName string
		githubURL   string
		localDir    string
		deploy      bool
		outputDir   string
		slug        string
	)

	cmd := &cobra.Command{
		Use:   "convert",
		Short: "Convert an NPX/stdio MCP server into a Cloudflare Worker",
		Long: `Converts a stdio-based MCP server package into an HTTP-compatible
Cloudflare Worker that can be deployed to Aerostack.

The converter:
  1. Downloads the npm package (or clones from GitHub)
  2. Analyzes the source for MCP tool definitions
  3. Detects environment variables (process.env.*)
  4. Generates a Worker wrapper with HTTP transport
  5. Generates a wrangler.toml
  6. Flags incompatible Node.js APIs

Examples:
  aerostack mcp convert --package @notionhq/notion-mcp-server
  aerostack mcp convert --package @linear/mcp-server --deploy
  aerostack mcp convert --github https://github.com/org/mcp-server
  aerostack mcp convert --dir ./my-local-mcp`,
		RunE: func(cmd *cobra.Command, args []string) error {
			printer.Header("MCP Server Converter")

			// Validate input — exactly one source
			sources := 0
			if packageName != "" {
				sources++
			}
			if githubURL != "" {
				sources++
			}
			if localDir != "" {
				sources++
			}
			if sources == 0 {
				return fmt.Errorf("specify a source: --package, --github, or --dir")
			}
			if sources > 1 {
				return fmt.Errorf("specify only one source: --package, --github, or --dir")
			}

			// 1. Fetch source
			var projectDir string
			var cleanup bool

			if packageName != "" {
				printer.Step("Downloading %s from npm...", packageName)
				dir, err := mcpconvert.FetchNPMPackage(packageName)
				if err != nil {
					return fmt.Errorf("fetch failed: %w", err)
				}
				projectDir = dir
				cleanup = true
				printer.Success("Downloaded to %s", dir)
			} else if githubURL != "" {
				printer.Step("Cloning %s...", githubURL)
				dir, err := mcpconvert.FetchGitHubRepo(githubURL)
				if err != nil {
					return fmt.Errorf("clone failed: %w", err)
				}
				projectDir = dir
				cleanup = true
				printer.Success("Cloned to %s", dir)
			} else {
				// Local directory
				absDir, err := filepath.Abs(localDir)
				if err != nil {
					return fmt.Errorf("resolve path: %w", err)
				}
				if _, err := os.Stat(filepath.Join(absDir, "package.json")); os.IsNotExist(err) {
					return fmt.Errorf("no package.json found in %s", absDir)
				}
				projectDir = absDir
			}

			if cleanup {
				defer mcpconvert.Cleanup(projectDir)
			}

			// 2. Analyze
			printer.Step("Analyzing MCP server...")
			analysis, err := mcpconvert.Analyze(projectDir)
			if err != nil {
				return fmt.Errorf("analysis failed: %w", err)
			}

			fmt.Println(printer.KeyVal("Package", analysis.PackageName))
			fmt.Println(printer.KeyVal("Entry", analysis.EntryPoint))
			fmt.Println(printer.KeyVal("Tools", fmt.Sprintf("%d detected", analysis.ToolCount)))
			fmt.Println(printer.KeyVal("Transport", transportDesc(analysis)))
			if len(analysis.EnvVars) > 0 {
				fmt.Println(printer.KeyVal("Env vars", strings.Join(analysis.EnvVars, ", ")))
			}

			// 3. Transform
			printer.Step("Generating Worker code...")
			result, err := mcpconvert.Transform(analysis)
			if err != nil {
				return fmt.Errorf("transform failed: %w", err)
			}

			// Show warnings
			for _, w := range result.Warnings {
				if strings.HasPrefix(w, "[ERROR]") {
					printer.Error("%s", w)
				} else if strings.HasPrefix(w, "[WARN]") {
					printer.Warn("%s", w)
				} else {
					printer.Hint("%s", w)
				}
			}

			// Check for blocking errors
			hasErrors := false
			for _, w := range result.Warnings {
				if strings.HasPrefix(w, "[ERROR]") {
					hasErrors = true
				}
			}
			if hasErrors {
				printer.Error("Conversion has blocking incompatibilities. Manual port required for flagged APIs.")
				printer.Hint("You can still generate the Worker code — flagged tools will fail at runtime.")
			}

			// 4. Write output
			if outputDir == "" {
				outputDir = "."
			}
			outPath := filepath.Join(outputDir, "converted-mcp-worker")
			os.MkdirAll(filepath.Join(outPath, "src"), 0755)

			// Write Worker code
			workerPath := filepath.Join(outPath, "src", "index.ts")
			if err := os.WriteFile(workerPath, []byte(result.WorkerCode), 0644); err != nil {
				return fmt.Errorf("write worker: %w", err)
			}

			// Write wrangler.toml
			wranglerPath := filepath.Join(outPath, "wrangler.toml")
			if err := os.WriteFile(wranglerPath, []byte(result.WranglerToml), 0644); err != nil {
				return fmt.Errorf("write wrangler.toml: %w", err)
			}

			// Write package.json
			pkgJSON := fmt.Sprintf(`{
  "name": "%s-worker",
  "version": "1.0.0",
  "private": true,
  "scripts": {
    "dev": "wrangler dev",
    "deploy": "wrangler deploy"
  },
  "dependencies": {
    "@modelcontextprotocol/sdk": "latest",
    "zod": "latest"
  },
  "devDependencies": {
    "wrangler": "latest",
    "typescript": "latest",
    "@cloudflare/workers-types": "latest"
  }
}`, mcpconvert.SanitizeSlug(analysis.PackageName))
			pkgPath := filepath.Join(outPath, "package.json")
			if err := os.WriteFile(pkgPath, []byte(pkgJSON), 0644); err != nil {
				return fmt.Errorf("write package.json: %w", err)
			}

			fmt.Println()
			printer.Success("Worker generated at: %s", outPath)
			fmt.Println()
			fmt.Println(printer.KeyVal("Worker", workerPath))
			fmt.Println(printer.KeyVal("Config", wranglerPath))
			fmt.Println(printer.KeyVal("Secrets", fmt.Sprintf("%d env vars to configure", len(result.DetectedSecrets))))
			fmt.Println()

			if len(result.DetectedSecrets) > 0 {
				printer.Hint("Required secrets (add to workspace or wrangler.toml):")
				for _, s := range result.DetectedSecrets {
					fmt.Printf("  %s\n", s)
				}
				fmt.Println()
			}

			// 5. Optional deploy
			if deploy {
				printer.Step("Deploying to Aerostack...")

				apiKey := credentials.GetAPIKey()
				if apiKey == "" {
					return fmt.Errorf("not logged in. Run 'aerostack login' first, or deploy manually with 'aerostack deploy mcp'")
				}

				// Build with esbuild first
				printer.Step("Bundling Worker...")
				bundledPath, err := mcpconvert.BundleWorker(outPath)
				if err != nil {
					printer.Warn("Bundle failed: %v", err)
					printer.Hint("Install dependencies first: cd %s && npm install && aerostack deploy mcp", outPath)
					return nil
				}

				if slug == "" {
					slug = mcpconvert.SanitizeSlug(analysis.PackageName)
				}

				resp, err := api.CommunityDeployMcp(apiKey, bundledPath, slug, "production")
				if err != nil {
					return fmt.Errorf("deploy failed: %w", err)
				}

				fmt.Println()
				printer.Success("MCP Server deployed!")
				fmt.Println(printer.KeyVal("URL", resp.WorkerURL))
				fmt.Println(printer.KeyVal("Slug", resp.Slug))
			} else {
				printer.Hint("Next steps:")
				fmt.Printf("  1. cd %s\n", outPath)
				fmt.Println("  2. Copy tool definitions from the original source into src/index.ts")
				fmt.Println("  3. npm install")
				fmt.Println("  4. wrangler dev  (test locally)")
				fmt.Println("  5. aerostack deploy mcp  (deploy to Aerostack)")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&packageName, "package", "", "NPM package name (e.g. @notionhq/notion-mcp-server)")
	cmd.Flags().StringVar(&githubURL, "github", "", "GitHub repository URL")
	cmd.Flags().StringVar(&localDir, "dir", "", "Local directory containing the MCP server")
	cmd.Flags().BoolVar(&deploy, "deploy", false, "Deploy to Aerostack after conversion")
	cmd.Flags().StringVarP(&outputDir, "output", "o", ".", "Output directory for generated Worker")
	cmd.Flags().StringVar(&slug, "slug", "", "Server slug for deployment (default: derived from package name)")

	return cmd
}

// NewMcpCommand creates the 'aerostack mcp' parent command.
func NewMcpCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "MCP server management commands",
		Long:  "Commands for converting, deploying, and managing MCP servers on Aerostack.",
	}

	cmd.AddCommand(NewMcpConvertCommand())

	return cmd
}

func transportDesc(a *mcpconvert.AnalysisResult) string {
	parts := []string{}
	if a.HasStdio {
		parts = append(parts, "stdio (will convert to HTTP)")
	}
	if a.HasHTTP {
		parts = append(parts, "HTTP (already compatible)")
	}
	if len(parts) == 0 {
		return "unknown (no transport detected)"
	}
	return strings.Join(parts, " + ")
}
