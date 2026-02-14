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
	"github.com/spf13/cobra"
)

// NewDeployCommand creates the 'aerostack deploy' command
func NewDeployCommand() *cobra.Command {
	var environment string
	var allServices bool
	var useWrangler bool

	cmd := &cobra.Command{
		Use:   "deploy [service-name]",
		Short: "Deploy services to Aerostack cloud",
		Long: `Deploy your services to the Aerostack cloud infrastructure.

When linked and logged in, deploys to Aerostack. Otherwise uses wrangler (your Cloudflare account).
Use --wrangler to force wrangler deploy.

Example:
  aerostack deploy --env staging
  aerostack deploy --env production
  aerostack deploy --wrangler  # Use your own Cloudflare account`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var serviceName string
			if len(args) > 0 {
				serviceName = args[0]
			}
			return deployService(serviceName, environment, allServices, useWrangler)
		},
	}

	cmd.Flags().StringVarP(&environment, "env", "e", "staging", "Target environment (staging/production)")
	cmd.Flags().BoolVar(&allServices, "all", false, "Deploy all services")
	cmd.Flags().BoolVar(&useWrangler, "wrangler", false, "Use wrangler (your Cloudflare account) instead of Aerostack")

	return cmd
}

func deployService(service, env string, all bool, useWrangler bool) error {
	// Validate env
	if env != "staging" && env != "production" {
		return fmt.Errorf("invalid env %q: use staging or production", env)
	}

	// 1. Check aerostack.toml
	if _, err := os.Stat("aerostack.toml"); os.IsNotExist(err) {
		return fmt.Errorf("aerostack.toml not found. Run 'aerostack init' first")
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

	// 4. Try Aerostack deploy if logged in (and not --wrangler)
	// Account key: no link required. Project key: link required.
	if !useWrangler {
		cred, _ := credentials.Load()
		if cred != nil && cred.APIKey != "" {
			validateResp, err := api.Validate(cred.APIKey)
			if err == nil {
				if validateResp.KeyType == "account" {
					return deployToAerostack(cfg, env, cred.APIKey, "")
				}
				// Project key: require link
				projLink, _ := link.Load()
				if projLink == nil || projLink.ProjectID == "" {
					return fmt.Errorf("project key requires link. Run 'aerostack link %s'", validateResp.ProjectID)
				}
				if validateResp.ProjectID != projLink.ProjectID {
					return fmt.Errorf("API key is for project %s but directory is linked to %s. Run 'aerostack link %s'", validateResp.ProjectID, projLink.ProjectID, validateResp.ProjectID)
				}
				return deployToAerostack(cfg, env, cred.APIKey, projLink.ProjectID)
			}
		}
	}

	// 5. Fall back to wrangler
	wranglerPath := "wrangler.toml"
	if err := devserver.GenerateWranglerToml(cfg, wranglerPath); err != nil {
		return fmt.Errorf("failed to generate wrangler.toml: %w", err)
	}

	if all {
		fmt.Printf("🚀 Deploying all services to %s (wrangler)...\n", env)
	}
	fmt.Printf("🚀 Deploying to %s (wrangler)...\n", env)

	if err := devserver.RunWranglerDeploy(wranglerPath, env); err != nil {
		return err
	}

	fmt.Println("\n✅ Deployment successful!")
	return nil
}

func deployToAerostack(cfg *devserver.AerostackConfig, env string, apiKey string, projectID string) error {
	// projectID empty = account key (use name from cfg). projectID set = project key (link verified by caller).
	projectName := ""
	if projectID == "" {
		projectName = cfg.Name
		if projectName == "" {
			projectName = "aerostack-app"
		}
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

	deployResp, err := api.Deploy(apiKey, workerPath, env, projectName)
	if err != nil {
		return err
	}

	fmt.Printf("\n✅ Deployed to Aerostack!\n")
	fmt.Printf("   URL: %s\n", deployResp.URL)
	fmt.Printf("   Project: %s (%s)\n", deployResp.Project.Name, deployResp.Project.Slug)
	return nil
}
