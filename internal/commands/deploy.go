package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/aerostackdev/cli/internal/agent"
	"github.com/aerostackdev/cli/internal/api"
	"github.com/aerostackdev/cli/internal/credentials"
	"github.com/aerostackdev/cli/internal/devserver"
	"github.com/aerostackdev/cli/internal/link"
	"github.com/aerostackdev/cli/internal/modules/deploy"
	"github.com/aerostackdev/cli/internal/pkg"
	"github.com/aerostackdev/cli/internal/provision"
	"github.com/spf13/cobra"
)

// NewDeployCommand creates the 'aerostack deploy' command
func NewDeployCommand() *cobra.Command {
	var environment string
	var allServices bool
	var ownAccount bool
	var isPublic bool
	var isPrivate bool

	cmd := &cobra.Command{
		Use:   "deploy [service-name]",
		Short: "Deploy services to Aerostack cloud",
		Long: `Deploy your services with two clear options:

  Default (Aerostack):  aerostack deploy
    Deploys to Aerostack's infrastructure. Requires login.
    Run 'aerostack login' first. Shows in admin Custom Logic.

  Your own account:     aerostack deploy --cloudflare
    Deploys to your Cloudflare account via wrangler.
    Requires: npx wrangler login (or CLOUDFLARE_API_TOKEN)

Examples:
  aerostack deploy --env staging
  aerostack deploy --env production
  aerostack deploy --cloudflare --env staging`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Initialize Agent for Deploy logic
			cwd, _ := os.Getwd()
			store, err := pkg.NewStore(cwd)
			if err != nil {
				return fmt.Errorf("failed to open PKG: %w", err)
			}
			ag, err := agent.NewAgent(store, false)
			if err != nil {
				return fmt.Errorf("failed to init AI: %w", err)
			}

			deployer := deploy.NewDeployAgent(ag)

			// 1. Pre-flight
			if err := deployer.PreCheck(cmd.Context()); err != nil {
				fmt.Printf("⚠️  Pre-flight check skipped: %v\n", err)
			}

			fmt.Println("🚀 Deploying project...")

			var serviceName string
			if len(args) > 0 {
				serviceName = args[0]
			}

			// 2. Actual Deploy Logic
			if err := deployService(serviceName, environment, allServices, ownAccount, isPublic, isPrivate); err != nil {
				importStrings := true // just a flag to know we might need strings package
				_ = importStrings

				fmt.Printf("\nDeployment Failed! Error details:\n%v\n\n", err)

				// 3. Failure Analysis
				// Skip AI analysis for obvious auth errors
				errStr := err.Error()
				if !strings.Contains(errStr, "API key invalid") && !strings.Contains(errStr, "not logged in") && !strings.Contains(errStr, "401") {
					_ = deployer.AnalyzeFailure(cmd.Context(), err)
				}
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&environment, "env", "e", "staging", "Target environment (staging/production)")
	cmd.Flags().BoolVar(&allServices, "all", false, "Deploy all services")
	cmd.Flags().BoolVar(&ownAccount, "cloudflare", false, "Deploy to your Cloudflare account (default: Aerostack)")
	cmd.Flags().BoolVar(&isPublic, "public", false, "Make the deployed service publicly accessible")
	cmd.Flags().BoolVar(&isPrivate, "private", false, "Make the deployed service private (requires authentication)")

	return cmd
}

func deployService(service, env string, all bool, ownAccount bool, isPublic bool, isPrivate bool) error {
	// Validate env
	if env != "staging" && env != "production" {
		return fmt.Errorf("invalid env %q: use staging or production", env)
	}

	// 1. Check aerostack.toml
	if _, err := os.Stat("aerostack.toml"); os.IsNotExist(err) {
		return fmt.Errorf("aerostack deploy requires you to have an aerostack.toml file in your project")
	}

	// Validate flags
	if isPublic && isPrivate {
		return fmt.Errorf("cannot specify both --public and --private flags")
	}

	// 2. Check Node.js
	if _, err := devserver.CheckNode(); err != nil {
		return err
	}

	// 3. Parse config
	cfg, err := devserver.ParseAerostackToml("aerostack.toml")
	if err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}
	devserver.EnsureDefaultD1(cfg)

	// Path 1: Deploy to user's own Cloudflare account (--cloudflare)
	if ownAccount {
		projectRoot, _ := os.Getwd()
		if projectRoot == "" {
			projectRoot = "."
		}
		fmt.Println("🔍 Checking resources in aerostack.toml...")
		if err := provision.ProvisionCloudflareResources(cfg, env, projectRoot); err != nil {
			return fmt.Errorf("provision resources: %w", err)
		}
		// Re-parse config in case provision updated aerostack.toml
		cfg, err = devserver.ParseAerostackToml("aerostack.toml")
		if err != nil {
			return fmt.Errorf("failed to re-parse config: %w", err)
		}
		wranglerPath := "wrangler.toml"
		if err := devserver.GenerateWranglerToml(cfg, wranglerPath); err != nil {
			return fmt.Errorf("failed to generate wrangler.toml: %w", err)
		}
		fmt.Printf("\n🚀 Deploying to your Cloudflare account (%s)...\n", env)
		if err := devserver.RunWranglerDeploy(wranglerPath, env); err != nil {
			return err
		}
		fmt.Println("\n✅ Deployment successful!")
		return nil
	}

	// Path 2: Deploy to Aerostack (default). Requires login.
	cred, _ := credentials.Load()
	if cred == nil || cred.APIKey == "" {
		return fmt.Errorf("not logged in. Run 'aerostack login' to deploy to Aerostack.\nTo deploy to your Cloudflare account instead, use: aerostack deploy --cloudflare")
	}
	validateResp, err := api.Validate(cred.APIKey)
	if err != nil {
		return fmt.Errorf("API key invalid or unreachable: %w\nRun 'aerostack login' to fix. Or use --cloudflare to deploy to your Cloudflare account.", err)
	}

	// Case A: Project Key (scoped to single project)
	if validateResp.KeyType == "project" {
		fmt.Printf("Authenticated as project: %s\n", validateResp.ProjectName)
		// Bypass link check, use project ID from key
		return deployToAerostack(cfg, env, cred.APIKey, validateResp.ProjectID, service, isPublic, isPrivate)
	}

	// Case B: Account Key (root access)
	// 1. Check if linked
	projLink, _ := link.Load()
	if projLink != nil && projLink.ProjectID != "" {
		return deployToAerostack(cfg, env, cred.APIKey, projLink.ProjectID, service, isPublic, isPrivate)
	}

	// 2. Not linked: Auto-create or find project by name
	projName := cfg.Name
	if projName == "" {
		projName = filepath.Base(filepath.Dir(".")) // fallback to folder name? or just error?
		if projName == "." || projName == "/" {
			return fmt.Errorf("project name missing in aerostack.toml")
		}
	}

	fmt.Printf("🔍 Checking project '%s'...\n", projName)
	projectMeta, err := api.GetProjectMetadata(cred.APIKey, projName)
	var projectID string

	if err == nil && projectMeta != nil {
		fmt.Printf("✅ Found existing project: %s (%s)\n", projectMeta.Name, projectMeta.ProjectID)
		projectID = projectMeta.ProjectID
	} else {
		// Assume 404/error means not found -> Create
		fmt.Printf("🆕 Project '%s' not found. Creating...\n", projName)
		createResp, err := api.CreateProject(cred.APIKey, projName)
		if err != nil {
			return fmt.Errorf("failed to create project '%s': %w", projName, err)
		}
		fmt.Printf("✅ Created project: %s (%s)\n", createResp.Name, createResp.ID)
		projectID = createResp.ID

		// Auto-link for future
		if err := link.Save(projectID); err == nil {
			fmt.Printf("🔗 Linked local directory to project %s\n", projectID)
		}
	}

	return deployToAerostack(cfg, env, cred.APIKey, projectID, service, isPublic, isPrivate)
}

func deployToAerostack(cfg *devserver.AerostackConfig, env string, apiKey string, projectID string, serviceName string, isPublic bool, isPrivate bool) error {
	// If serviceName is empty (no CLI arg), use name from aerostack.toml
	if serviceName == "" {
		serviceName = cfg.Name
	}
	if serviceName == "" {
		serviceName = "default"
	}

	fmt.Printf("🚀 Deploying to Aerostack (%s)...\n", env)

	// Build worker with esbuild
	mainEntry := cfg.Main
	if mainEntry == "" {
		mainEntry = "src/index.ts"
	}

	// Clean dist directory to avoid stale builds
	os.RemoveAll("dist")
	os.MkdirAll("dist", 0755)

	// Check for nodejs_compat_v2
	hasNodeCompatV2 := false
	for _, flag := range cfg.CompatibilityFlags {
		if flag == "nodejs_compat_v2" {
			hasNodeCompatV2 = true
			break
		}
	}

	// Create node-mocks for both legacy and nodejs_compat_v2 paths if they load modules like express
	mockDir := ".aerostack"
	os.MkdirAll(mockDir, 0755)
	mockPath := filepath.Join(mockDir, "node-mocks.cjs")
	// Minimal mock: exports most things as empty objects or dummy functions
	mockContent := `
const noop = () => {};
const emptyObj = {};
const handler = {
	get: (target, prop) => {
		if (prop === 'prototype') return emptyObj;
		if (prop === 'on' || prop === 'once' || prop === 'emit') return noop;
		return proxy;
	},
	construct: () => proxy,
	apply: () => proxy
};
const proxy = new Proxy(noop, handler);

Object.assign(proxy, { 
	isatty: () => false,
	createServer: () => ({ listen: () => ({ on: () => {} }), on: () => {} }),
	readFileSync: () => { throw new Error("fs.readFileSync is not supported in Workers. Use cache or bindings.") },
	METHODS: ["GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"],
	STATUS_CODES: { 200: "OK", 404: "Not Found", 500: "Internal Server Error" },
	IncomingMessage: proxy,
	ServerResponse: proxy,
	Agent: proxy,
	Socket: proxy,
	networkInterfaces: () => ({}),
	arch: () => "arm64",
	platform: () => "linux"
});

module.exports = proxy;
`
	os.WriteFile(mockPath, []byte(mockContent), 0644)

	// esbuild requires relative paths to start with ./ or ../ or it treats them as bare modules
	mockPathForEsbuild := "./" + filepath.ToSlash(filepath.Join(".aerostack", "node-mocks.cjs"))

	args := []string{
		"--yes",
		"esbuild",
		mainEntry,
		"--bundle",
		"--outfile=dist/worker.js",
		"--format=esm",
		"--alias:@shared=./shared",
		"--minify",
		"--external:node:*",
		"--external:cloudflare:*",
	}

	// Add banner with createRequire shim and prototype fixes
	banner := "import { createRequire } from 'node:module'; const require = createRequire('/'); if (typeof Buffer !== 'undefined' && !Buffer.prototype.hasOwnProperty) Buffer.prototype.hasOwnProperty = Object.prototype.hasOwnProperty;"
	args = append(args, "--banner:js="+banner)

	// Add aliases for Node.js modules
	nodeModules := []string{"path", "url", "crypto", "events", "util", "stream", "buffer", "string_decoder", "assert", "timers", "async_hooks", "console", "querystring", "zlib", "punycode"}

	if hasNodeCompatV2 {
		for _, m := range nodeModules {
			args = append(args, fmt.Sprintf("--alias:%s=node:%s", m, m))
		}
	} else {
		// Legacy / Default
		args = append(args, "--platform=node")
		for _, m := range nodeModules {
			args = append(args, fmt.Sprintf("--alias:%s=node:%s", m, m))
		}
	}

	// Mock common unsupported modules
	unsupported := []string{"fs", "os", "http", "https", "net", "dns", "tty", "tls", "child_process"}
	for _, m := range unsupported {
		args = append(args, fmt.Sprintf("--alias:%s=%s", m, mockPathForEsbuild))
		args = append(args, fmt.Sprintf("--alias:node:%s=%s", m, mockPathForEsbuild))
	}

	fmt.Printf("📦 Bundling %s...\n", mainEntry)

	cmd := exec.Command("npx", args...)
	cmd.Dir = "."
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	distPath := "dist"
	files := make(map[string]string)
	entries, err := os.ReadDir(distPath)
	if err != nil {
		return fmt.Errorf("failed to read dist directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			name := entry.Name()
			// We only upload .js, .mjs, .wasm, etc. as modules.
			// Cloudflare Workers expect the main entry as worker.js.
			localPath := filepath.Join(distPath, name)
			if name == "worker.js" {
				files["worker"] = localPath
			} else {
				// Upload other files as modules
				files[name] = localPath
			}
		}
	}

	if _, ok := files["worker"]; !ok {
		return fmt.Errorf("build output worker.js not found in %s", distPath)
	}

	// Prepare bindings JSON
	// Combine default bindings with env overrides
	d1s := cfg.D1Databases
	if overrides, ok := cfg.EnvOverrides[env]; ok && len(overrides) > 0 {
		d1s = overrides
	}

	// Filter out stub/local-only bindings that only exist for local dev
	// (e.g. 'local-queue', 'local-kv') — these are not real Cloudflare resources
	var realQueues []devserver.Queue
	for _, q := range cfg.Queues {
		if !strings.HasPrefix(q.Name, "local-") {
			realQueues = append(realQueues, q)
		}
	}
	var realKVNamespaces []devserver.KVNamespace
	for _, ns := range cfg.KVNamespaces {
		if !strings.HasPrefix(ns.ID, "local-") {
			realKVNamespaces = append(realKVNamespaces, ns)
		}
	}
	var realD1s []devserver.D1Database
	for _, db := range d1s {
		if db.DatabaseID != "aerostack-local" && !strings.HasPrefix(db.DatabaseID, "local-") {
			realD1s = append(realD1s, db)
		}
	}

	type BindingPayload struct {
		D1Databases  []devserver.D1Database  `json:"d1_databases,omitempty"`
		KVNamespaces []devserver.KVNamespace `json:"kv_namespaces,omitempty"`
		Queues       []devserver.Queue       `json:"queues,omitempty"`
		AI           bool                    `json:"ai,omitempty"`
	}

	bindingsPayload := BindingPayload{
		D1Databases:  realD1s,
		KVNamespaces: realKVNamespaces,
		Queues:       realQueues,
		AI:           cfg.AI,
	}

	bData, _ := json.Marshal(bindingsPayload)
	bindingsJSON := string(bData)

	deployResp, err := api.Deploy(apiKey, files, env, serviceName, isPublic, isPrivate, bindingsJSON, cfg.CompatibilityDate, cfg.CompatibilityFlags)
	if err != nil {
		return err
	}

	fmt.Printf("\n✅ Deployed to Aerostack!\n")
	fmt.Printf("   URL: %s\n", deployResp.PublicURL)

	if deployResp.Project.Slug != "" {
		fmt.Printf("   Dashboard: %s\n", deployResp.PublicURL)
	}

	fmt.Printf("\n🔒 Authentication Status:\n")
	if deployResp.IsPublic {
		fmt.Printf("   Your API is PUBLIC. Anyone can access it without an API key.\n")
		fmt.Printf("   To make it private, deploy with: aerostack deploy --private\n")
		fmt.Printf("\n   Test it now:\n")
		fmt.Printf("   curl %s\n", deployResp.PublicURL)
	} else {
		fmt.Printf("   Your API is PRIVATE. Requests require your Project API Key.\n")
		fmt.Printf("   To make it public, deploy with: aerostack deploy --public\n")
		fmt.Printf("\n   Test it now (using query param):\n")
		fmt.Printf("   curl \"%s?apiKey=%s\"\n", deployResp.PublicURL, apiKey)
		fmt.Printf("\n   Test it now (using header):\n")
		fmt.Printf("   curl -H \"X-API-Key: %s\" %s\n", apiKey, deployResp.PublicURL)
	}

	return nil
}
