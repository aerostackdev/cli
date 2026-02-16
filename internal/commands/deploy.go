package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/aerostackdev/cli/internal/api"
	"github.com/aerostackdev/cli/internal/credentials"
	"github.com/aerostackdev/cli/internal/devserver"
	"github.com/aerostackdev/cli/internal/link"
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
			var serviceName string
			if len(args) > 0 {
				serviceName = args[0]
			}
			return deployService(serviceName, environment, allServices, ownAccount, isPublic, isPrivate)
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
	buildCmd := fmt.Sprintf("npx esbuild %q --bundle --outfile=dist/worker.js --format=esm --alias:@shared=./shared", mainEntry)

	cmd := exec.Command("sh", "-c", buildCmd)
	cmd.Dir = "."
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	workerPath := filepath.Join("dist", "worker.js")
	if _, err := os.Stat(workerPath); err != nil {
		return fmt.Errorf("build output not found at %s: %w", workerPath, err)
	}

	deployResp, err := api.Deploy(apiKey, workerPath, env, serviceName, isPublic, isPrivate)
	if err != nil {
		return err
	}

	fmt.Printf("\n✅ Deployed to Aerostack!\n")
	fmt.Printf("   URL: %s\n", deployResp.PublicURL)

	if deployResp.Project.Slug != "" {
		fmt.Printf("   Dashboard: https://aerocall.ai/project/%s\n", deployResp.Project.Slug)
	}
	return nil
}
