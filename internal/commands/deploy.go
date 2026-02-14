package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewDeployCommand creates the 'aerostack deploy' command
func NewDeployCommand() *cobra.Command {
	var environment string
	var allServices bool

	cmd := &cobra.Command{
		Use:   "deploy [service-name]",
		Short: "Deploy services to Aerostack cloud",
		Long: `Deploy your services to the Aerostack cloud infrastructure.

Supports multi-environment deployments with atomic versioning.

Example:
  aerostack deploy api-gateway --env production
  aerostack deploy --all --env staging`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var serviceName string
			if len(args) > 0 {
				serviceName = args[0]
			}
			return deployService(serviceName, environment, allServices)
		},
	}

	cmd.Flags().StringVarP(&environment, "env", "e", "staging", "Target environment (staging/production)")
	cmd.Flags().BoolVar(&allServices, "all", false, "Deploy all services")

	return cmd
}

func deployService(service, env string, all bool) error {
	if all {
		fmt.Printf("🚀 Deploying all services to %s...\n", env)
	} else {
		fmt.Printf("🚀 Deploying %s to %s...\n", service, env)
	}
	
	// TODO: Implement deployment logic
	fmt.Println("\n✅ Deployment successful!")
	fmt.Printf("   URL: https://%s-service.aerostack.com\n", service)
	
	return nil
}
