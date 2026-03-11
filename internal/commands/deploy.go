package commands

import (
	"bufio"
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
	"github.com/aerostackdev/cli/internal/printer"
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
	var syncSecrets bool

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
				printer.Warn("Pre-flight check skipped: %v", err)
			}

			printer.Header("Deploying Project")

			var serviceName string
			if len(args) > 0 {
				serviceName = args[0]
			}

			// 2. Actual Deploy Logic
			if err := deployService(serviceName, environment, allServices, ownAccount, isPublic, isPrivate, syncSecrets); err != nil {
				importStrings := true // just a flag to know we might need strings package
				_ = importStrings

				fmt.Println()
				printer.Error("Deployment Failed! Error details:\n%v", err)
				fmt.Println()

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
	cmd.Flags().BoolVar(&syncSecrets, "sync-secrets", false, "Push non-standard .dev.vars keys as secrets to the target environment before deploying")

	// Subcommands
	cmd.AddCommand(NewDeployMcpCommand())
	cmd.AddCommand(NewDeploySkillCommand())

	return cmd
}

func deployService(service, env string, all bool, ownAccount bool, isPublic bool, isPrivate bool, syncSecrets bool) error {
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

	// Push .dev.vars secrets to the target environment before deploying (--sync-secrets)
	if syncSecrets {
		if err := syncSecretsFromDevVars(env); err != nil {
			printer.Warn("Failed to sync secrets: %v", err)
		}
	}

	// Path 1: Deploy to user's own Cloudflare account (--cloudflare)
	if ownAccount {
		projectRoot, _ := os.Getwd()
		if projectRoot == "" {
			projectRoot = "."
		}
		printer.Step("Checking resources in aerostack.toml...")
		if err := provision.ProvisionCloudflareResources(cfg, env, projectRoot); err != nil {
			return fmt.Errorf("provision resources: %w", err)
		}
		// Re-parse config in case provision updated aerostack.toml
		cfg, err = devserver.ParseAerostackToml("aerostack.toml")
		if err != nil {
			return fmt.Errorf("failed to re-parse config: %w", err)
		}
		// Keep generated wrangler.toml inside .aerostack/ to avoid cluttering project root
		if err := os.MkdirAll(".aerostack", 0755); err != nil {
			return fmt.Errorf("failed to create .aerostack directory: %w", err)
		}
		wranglerPath := filepath.Join(".aerostack", "wrangler.toml")
		if err := devserver.GenerateWranglerToml(cfg, wranglerPath); err != nil {
			return fmt.Errorf("failed to generate wrangler.toml: %w", err)
		}
		fmt.Println()
		printer.Step("Deploying to your Cloudflare account (%s)...", env)
		if err := devserver.RunWranglerDeploy(wranglerPath, env); err != nil {
			return err
		}
		fmt.Println()
		printer.Success("Deployment successful!")
		printSecretsReminder(cfg, env)
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
		printer.Step("Authenticated")
		fmt.Println(printer.KeyVal("Project scope", validateResp.ProjectName))
		// Bypass link check, use project ID from key
		return deployToAerostack(cfg, env, cred.APIKey, validateResp.ProjectID, service, isPublic, isPrivate)
	}

	// Case B: Account Key (root access)
	// Priority 1: explicit project_id in aerostack.toml
	if cfg.ProjectID != "" {
		printer.Step("Using project from aerostack.toml: %s", cfg.ProjectID)
		return deployToAerostack(cfg, env, cred.APIKey, cfg.ProjectID, service, isPublic, isPrivate)
	}

	// Priority 2: locally linked project (.aerostack/project.json)
	projLink, _ := link.Load()
	if projLink != nil && projLink.ProjectID != "" {
		return deployToAerostack(cfg, env, cred.APIKey, projLink.ProjectID, service, isPublic, isPrivate)
	}

	// Priority 3: Auto-create or find project by name
	projName := cfg.Name
	if projName == "" {
		projName = filepath.Base(filepath.Dir(".")) // fallback to folder name? or just error?
		if projName == "." || projName == "/" {
			return fmt.Errorf("project name missing in aerostack.toml")
		}
	}

	printer.Step("Checking project '%s'...", projName)
	projectMeta, err := api.GetProjectMetadata(cred.APIKey, projName)
	var projectID string

	if err == nil && projectMeta != nil {
		printer.Success("Found existing project: %s (%s)", projectMeta.Name, projectMeta.ProjectID)
		projectID = projectMeta.ProjectID
	} else {
		// Assume 404/error means not found -> Create
		printer.Step("Project '%s' not found. Creating...", projName)
		createResp, err := api.CreateProject(cred.APIKey, projName)
		if err != nil {
			return fmt.Errorf("failed to create project '%s': %w", projName, err)
		}
		printer.Success("Created project: %s (%s)", createResp.Name, createResp.ID)
		projectID = createResp.ID

		// Auto-link for future
		if err := link.Save(projectID); err == nil {
			printer.Hint("Linked local directory to project: %s", projectID)
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

	fmt.Println()
	printer.Step("Deploying to Aerostack (%s)...", env)

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

	printer.Step("Bundling %s...", mainEntry)

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

	deployResp, err := api.Deploy(apiKey, files, env, serviceName, projectID, isPublic, isPrivate, bindingsJSON, cfg.CompatibilityDate, cfg.CompatibilityFlags)
	if err != nil {
		return err
	}

	fmt.Println()
	printer.Success("Deployed to Aerostack!")
	fmt.Println(printer.KeyVal("URL", deployResp.PublicURL))
	printSecretsReminder(cfg, env)

	if deployResp.Project.Slug != "" {
		fmt.Println(printer.KeyVal("Dashboard", deployResp.PublicURL)) // TODO: switch to actual dashboard URL when ready
	}

	fmt.Println()
	printer.Step("Authentication Status")
	if deployResp.IsPublic {
		printer.Hint("Your API is PUBLIC. Anyone can access it without an API key.")
		printer.Hint("To make it private, deploy with: %s", printer.Command("aerostack deploy --private"))
		fmt.Println()
		printer.Hint("Test it now:")
		fmt.Println("  " + printer.Command(fmt.Sprintf("curl %s", deployResp.PublicURL)))
	} else {
		printer.Hint("Your API is PRIVATE. Requests require your Project API Key.")
		printer.Hint("To make it public, deploy with: %s", printer.Command("aerostack deploy --public"))
		fmt.Println()
		printer.Hint("Test it now (using query param):")
		fmt.Println("  " + printer.Command(fmt.Sprintf("curl \"%s?apiKey=%s\"", deployResp.PublicURL, apiKey)))
		fmt.Println()
		printer.Hint("Test it now (using header):")
		fmt.Println("  " + printer.Command(fmt.Sprintf("curl -H \"X-API-Key: %s\" %s", apiKey, deployResp.PublicURL)))
	}

	return nil
}

// standardAerostackKeys are keys managed by the Aerostack platform and should not be
// auto-pushed as user secrets.
var standardAerostackKeys = map[string]bool{
	"AEROSTACK_PROJECT_ID": true,
	"AEROSTACK_API_KEY":    true,
	"AEROSTACK_API_URL":    true,
}

// parseDevVars reads .dev.vars and returns a map of key→value pairs.
// Lines starting with # are treated as comments. Empty lines are ignored.
func parseDevVars(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	vars := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.IndexByte(line, '=')
		if idx <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		// Strip optional surrounding quotes
		if len(val) >= 2 && ((val[0] == '"' && val[len(val)-1] == '"') || (val[0] == '\'' && val[len(val)-1] == '\'')) {
			val = val[1 : len(val)-1]
		}
		vars[key] = val
	}
	return vars, scanner.Err()
}

// syncSecretsFromDevVars reads .dev.vars, filters out standard Aerostack keys, and pushes
// the remaining key/value pairs as Cloudflare secrets for the given environment.
func syncSecretsFromDevVars(env string) error {
	vars, err := parseDevVars(".dev.vars")
	if err != nil {
		if os.IsNotExist(err) {
			printer.Hint("No .dev.vars file found — skipping secret sync")
			return nil
		}
		return fmt.Errorf("reading .dev.vars: %w", err)
	}

	pushed := 0
	for key, val := range vars {
		if standardAerostackKeys[key] {
			continue
		}
		printer.Step("Pushing secret: %s → %s", key, env)
		if err := runSecretsSet(key, []string{key, val}, env); err != nil {
			printer.Warn("Failed to set %s: %v", key, err)
		} else {
			pushed++
		}
	}

	if pushed == 0 {
		printer.Hint("No user secrets found in .dev.vars to sync (only standard Aerostack keys were present)")
	}
	return nil
}

// printSecretsReminder prints a post-deploy reminder about secrets that may need to be set.
// It fires when the project uses Neon (postgres_databases) or when .dev.vars contains
// non-standard keys that have not been explicitly synced.
func printSecretsReminder(cfg *devserver.AerostackConfig, env string) {
	var reminders []string

	if len(cfg.PostgresDatabases) > 0 {
		reminders = append(reminders,
			fmt.Sprintf("  DATABASE_URL  →  aerostack secrets set DATABASE_URL \"<neon-url>\" --env %s", env),
		)
	}

	// Also check .dev.vars for any other non-standard keys
	if vars, err := parseDevVars(".dev.vars"); err == nil {
		for key := range vars {
			if standardAerostackKeys[key] {
				continue
			}
			// Skip DATABASE_URL if already covered above
			if key == "DATABASE_URL" && len(cfg.PostgresDatabases) > 0 {
				continue
			}
			reminders = append(reminders,
				fmt.Sprintf("  %s  →  aerostack secrets set %s \"<value>\" --env %s", key, key, env),
			)
		}
	}

	if len(reminders) == 0 {
		return
	}

	fmt.Println()
	printer.Warn("Production secrets needed — run the following commands (or use --sync-secrets on the next deploy):")
	for _, r := range reminders {
		fmt.Println(r)
	}
	fmt.Println()
}
