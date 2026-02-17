package deploy

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/aerostackdev/cli/internal/agent"
	"github.com/aerostackdev/cli/internal/modules/auth"
)

type DeployAgent struct {
	agent *agent.Agent
}

func NewDeployAgent(agent *agent.Agent) *DeployAgent {
	return &DeployAgent{agent: agent}
}

// PreCheck runs safety checks before deployment
func (d *DeployAgent) PreCheck(ctx context.Context) error {
	fmt.Println("🛫 Running Pre-flight Checks via AI Agent...")

	// 1. Re-use Auth Doctor
	// We verify that the user is authenticated and config is valid
	// In a real scenario, we might want a silent mode for Doctor
	doc := auth.NewDoctor(d.agent)
	// We'll trust the doctor to print output for now
	if err := doc.Run(ctx); err != nil {
		return fmt.Errorf("auth check failed: %w", err)
	}

	// 2. Check for uncommitted changes
	// Simple git check
	cmd := exec.Command("git", "status", "--porcelain")
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		fmt.Println("⚠️  Warning: You have uncommitted changes. Deploying might not reflect your current work.")
		// We don't block, just warn
	}

	fmt.Println("✅ Pre-flight checks passed.")
	return nil
}

// AnalyzeFailure asks the AI to explain why the deployment failed
func (d *DeployAgent) AnalyzeFailure(ctx context.Context, err error) error {
	fmt.Println("\n💥 Deployment Failed. Asking Aerostack AI for help...")

	prompt := fmt.Sprintf(`The 'aerostack deploy' command failed with the following error:
"%v"

You are the Deployment Specialist. 
Analyze this error. 
Is it a permissions issue? A syntax error? A missing configuration?
Suggest a specific fix command or code change.`, err)

	return d.agent.Resolve(ctx, prompt)
}
