package devserver

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// RunWranglerDeploy runs wrangler deploy for the given environment.
// Uses Aerostack-managed infra (wrangler deploys to configured Cloudflare account).
func RunWranglerDeploy(wranglerTomlPath string, env string) error {
	absPath, err := filepath.Abs(wranglerTomlPath)
	if err != nil {
		return err
	}
	projectRoot := filepath.Dir(absPath)

	args := []string{"-y", "wrangler@latest", "deploy", "--config", absPath}
	if env != "" {
		args = append(args, "--env", env)
	}

	cmd := exec.Command("npx", args...)
	cmd.Dir = projectRoot
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "NPX_UPDATE_NOTIFIER=false")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("wrangler deploy failed: %w", err)
	}
	return nil
}
