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
	cmd.AddCommand(NewMcpInitCommand())
	cmd.AddCommand(NewMcpPullCommand())

	return cmd
}

// NewMcpInitCommand creates 'aerostack mcp init <name>'.
func NewMcpInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init <name>",
		Short: "Scaffold a new MCP server directory",
		Long: `Creates a new MCP server directory with boilerplate files.

The server will be scaffolded at MCP/mcp-<name>/ (or mcp-<name>/ if already inside MCP/).

Examples:
  aerostack mcp init github
  aerostack mcp init my-tool`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := strings.ToLower(strings.TrimSpace(args[0]))
			name = strings.TrimPrefix(name, "mcp-")

			printer.Header("MCP Server Init")

			// Determine output directory: if cwd already ends in MCP/, write mcp-<name>/ here
			cwd, _ := os.Getwd()
			var outDir string
			if filepath.Base(cwd) == "MCP" {
				outDir = filepath.Join(cwd, "mcp-"+name)
			} else {
				outDir = filepath.Join(cwd, "MCP", "mcp-"+name)
			}

			printer.Step("Creating %s/", outDir)

			if err := os.MkdirAll(filepath.Join(outDir, "src"), 0755); err != nil {
				return fmt.Errorf("create dirs: %w", err)
			}

			// src/index.ts
			indexTs := fmt.Sprintf(`/**
 * %s MCP Server — Aerostack Worker
 * Edit the TOOLS array and callTool() to implement your server.
 */

const TOOLS = [
  {
    name: "example_tool",
    description: "An example tool — replace with your real tools.",
    inputSchema: {
      type: "object",
      properties: {
        input: { type: "string", description: "Input parameter" },
      },
      required: ["input"],
    },
  },
];

async function callTool(name: string, args: Record<string, unknown>, env: Record<string, string>): Promise<unknown> {
  // The Aerostack gateway injects secrets as X-Mcp-Secret-<ENV_VAR_NAME> headers.
  // Access them via env: e.g. env["API_KEY"] for X-Mcp-Secret-API-KEY header.
  const apiKey = env["API_KEY"] ?? "";

  switch (name) {
    case "example_tool": {
      const input = String(args.input ?? "");
      // TODO: implement your tool logic here
      return { result: "echo: " + input };
    }
    default:
      throw new Error("Unknown tool: " + name);
  }
}

// ─── JSON-RPC 2.0 handler ────────────────────────────────────────────────────

export default {
  async fetch(request: Request, rawEnv: unknown): Promise<Response> {
    const env = rawEnv as Record<string, string>;

    // Health check
    if (request.method === "GET") {
      return Response.json({ ok: true, server: "mcp-%s", tools: TOOLS.length });
    }

    // Collect secrets injected by Aerostack gateway
    const secrets: Record<string, string> = {};
    for (const [header, value] of request.headers.entries()) {
      const lower = header.toLowerCase();
      if (lower.startsWith("x-mcp-secret-")) {
        const envKey = lower.replace("x-mcp-secret-", "").toUpperCase().replace(/-/g, "_");
        secrets[envKey] = value;
      }
    }
    const mergedEnv = { ...env, ...secrets };

    let body: Record<string, unknown>;
    try {
      body = await request.json() as Record<string, unknown>;
    } catch {
      return Response.json({ jsonrpc: "2.0", error: { code: -32700, message: "Parse error" }, id: null }, { status: 400 });
    }

    const { jsonrpc, id, method, params } = body as {
      jsonrpc: string;
      id: unknown;
      method: string;
      params: Record<string, unknown>;
    };

    if (jsonrpc !== "2.0") {
      return Response.json({ jsonrpc: "2.0", error: { code: -32600, message: "Invalid Request" }, id: null }, { status: 400 });
    }

    try {
      switch (method) {
        case "initialize":
          return Response.json({
            jsonrpc: "2.0",
            id,
            result: {
              protocolVersion: "2024-11-05",
              capabilities: { tools: {} },
              serverInfo: { name: "mcp-%s", version: "1.0.0" },
            },
          });

        case "tools/list":
          return Response.json({ jsonrpc: "2.0", id, result: { tools: TOOLS } });

        case "tools/call": {
          const toolName = String((params as any)?.name ?? "");
          const toolArgs = ((params as any)?.arguments ?? {}) as Record<string, unknown>;
          const result = await callTool(toolName, toolArgs, mergedEnv);
          return Response.json({
            jsonrpc: "2.0",
            id,
            result: {
              content: [{ type: "text", text: typeof result === "string" ? result : JSON.stringify(result) }],
            },
          });
        }

        default:
          return Response.json({
            jsonrpc: "2.0",
            id,
            error: { code: -32601, message: "Method not found" },
          });
      }
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : String(err);
      return Response.json({
        jsonrpc: "2.0",
        id,
        error: { code: -32603, message: msg },
      });
    }
  },
};
`, name, name, name)

			// aerostack.toml
			aerostackToml := fmt.Sprintf(`name = "mcp-%s"
main = "src/index.ts"
compatibility_date = "2024-11-01"
compatibility_flags = ["nodejs_compat"]
`, name)

			// package.json
			packageJSON := fmt.Sprintf(`{
  "name": "aerostack-mcp-%s",
  "version": "1.0.0",
  "private": true,
  "scripts": {
    "dev": "wrangler dev",
    "build": "esbuild src/index.ts --bundle --outfile=dist/worker.js --format=esm --target=esnext --platform=neutral",
    "deploy": "aerostack deploy mcp %s"
  },
  "devDependencies": {
    "esbuild": "latest",
    "typescript": "latest",
    "wrangler": "latest",
    "@cloudflare/workers-types": "latest"
  }
}
`, name, name)

			// tsconfig.json
			tsconfig := `{
  "compilerOptions": {
    "target": "ES2022",
    "module": "ES2022",
    "moduleResolution": "bundler",
    "lib": ["ES2022"],
    "types": ["@cloudflare/workers-types"],
    "strict": true,
    "noEmit": true
  },
  "include": ["src/**/*.ts"]
}
`

			files := map[string]string{
				filepath.Join("src", "index.ts"): indexTs,
				"aerostack.toml":                 aerostackToml,
				"package.json":                   packageJSON,
				"tsconfig.json":                  tsconfig,
			}

			for rel, content := range files {
				dest := filepath.Join(outDir, rel)
				if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
					return fmt.Errorf("create dir for %s: %w", rel, err)
				}
				if err := os.WriteFile(dest, []byte(content), 0644); err != nil {
					return fmt.Errorf("write %s: %w", rel, err)
				}
				fmt.Printf("  created %s\n", filepath.Join("MCP", "mcp-"+name, rel))
			}

			fmt.Println()
			printer.Success("MCP server scaffolded at %s", outDir)
			fmt.Println()
			printer.Hint("Next steps:")
			fmt.Printf("  1. Edit %s\n", filepath.Join(outDir, "src", "index.ts"))
			fmt.Printf("  2. aerostack deploy mcp %s\n", name)
			return nil
		},
	}
}

// NewMcpPullCommand creates 'aerostack mcp pull <slug>'.
func NewMcpPullCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "pull <slug>",
		Short: "Pull an MCP server from Aerostack to your local directory",
		Long: `Downloads an MCP server's source files from Aerostack and writes them locally.

Examples:
  aerostack mcp pull mcp-github            (pulls your own @you/mcp-github)
  aerostack mcp pull @johndoe/mcp-github   (pulls another developer's server)`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			slug := args[0]

			apiKey := credentials.GetAPIKey()
			if apiKey == "" {
				return fmt.Errorf("not logged in. Run: aerostack login")
			}

			// If no @ prefix, resolve to own scoped slug via the API
			scopedSlug := slug
			if !strings.HasPrefix(scopedSlug, "@") {
				// Let the API prefix with the authenticated user's username
				baseName := strings.TrimPrefix(scopedSlug, "mcp-")
				scopedSlug = "mcp-" + baseName
			}

			printer.Step("Pulling %s...", scopedSlug)

			resp, err := api.McpPull(apiKey, scopedSlug)
			if err != nil {
				return err
			}

			// Derive local directory name from slug
			// @username/mcp-name → name
			baseName := resp.Slug
			if idx := strings.LastIndex(baseName, "/"); idx >= 0 {
				baseName = baseName[idx+1:]
			}
			baseName = strings.TrimPrefix(baseName, "mcp-")

			cwd, _ := os.Getwd()
			var outDir string
			if filepath.Base(cwd) == "MCP" {
				outDir = filepath.Join(cwd, "mcp-"+baseName)
			} else {
				outDir = filepath.Join(cwd, "MCP", "mcp-"+baseName)
			}

			if err := os.MkdirAll(filepath.Join(outDir, "src"), 0755); err != nil {
				return fmt.Errorf("create dirs: %w", err)
			}

			files := map[string]string{
				filepath.Join("src", "index.ts"): resp.SrcIndexTs,
				"aerostack.toml":                 resp.AerostackToml,
				"package.json":                   resp.PackageJson,
			}

			for rel, content := range files {
				dest := filepath.Join(outDir, rel)
				if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
					return fmt.Errorf("create dir for %s: %w", rel, err)
				}
				if err := os.WriteFile(dest, []byte(content), 0644); err != nil {
					return fmt.Errorf("write %s: %w", rel, err)
				}
			}

			fmt.Println()
			printer.Success("Pulled to %s", outDir)
			printer.Hint("Edit files, then run: aerostack deploy mcp %s", baseName)
			return nil
		},
	}
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
