package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/aerostackdev/cli/internal/devserver"
	"github.com/spf13/cobra"
)

// NewSecretsCommand creates the 'aerostack secrets' command
func NewSecretsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "secrets",
		Short: "Manage secrets for staging and production",
		Long: `Manage Cloudflare Workers secrets.
Local dev: use .dev.vars (never commit). Staging/Prod: use these commands.

  aerostack secrets list [--env staging|production]
  aerostack secrets set KEY value [--env staging|production]`,
	}

	cmd.AddCommand(NewSecretsListCommand())
	cmd.AddCommand(NewSecretsSetCommand())
	return cmd
}

// NewSecretsListCommand creates 'aerostack secrets list'
func NewSecretsListCommand() *cobra.Command {
	var env string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List secret names (values are never shown)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSecretsList(env)
		},
	}
	cmd.Flags().StringVar(&env, "env", "", "Environment (staging/production)")
	return cmd
}

// NewSecretsSetCommand creates 'aerostack secrets set KEY value'
func NewSecretsSetCommand() *cobra.Command {
	var env string

	cmd := &cobra.Command{
		Use:   "set [key] [value]",
		Short: "Set a secret for the given environment",
		Long: `Set a secret. Value can be passed as argument or via stdin.
  aerostack secrets set API_KEY my-secret --env production
  echo -n "my-secret" | aerostack secrets set API_KEY --env production`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSecretsSet(args[0], args, env)
		},
	}
	cmd.Flags().StringVar(&env, "env", "production", "Environment (staging/production)")
	return cmd
}

func ensureWranglerToml() error {
	if _, err := os.Stat("wrangler.toml"); err == nil {
		return nil
	}
	cfg, err := devserver.ParseAerostackToml("aerostack.toml")
	if err != nil {
		return fmt.Errorf("failed to parse aerostack.toml: %w", err)
	}
	devserver.EnsureDefaultD1(cfg)
	return devserver.GenerateWranglerToml(cfg, "wrangler.toml")
}

func runSecretsList(env string) error {
	if _, err := os.Stat("aerostack.toml"); os.IsNotExist(err) {
		return fmt.Errorf("aerostack.toml not found. Run 'aerostack init' first")
	}
	if err := ensureWranglerToml(); err != nil {
		return err
	}

	args := []string{"-y", "wrangler@latest", "secret", "list", "--config", "wrangler.toml"}
	if env != "" {
		args = append(args, "--env", env)
	}

	c := exec.Command("npx", args...)
	c.Dir = "."
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Env = append(os.Environ(), "NPX_UPDATE_NOTIFIER=false")

	absDir, _ := filepath.Abs(".")
	c.Dir = absDir

	if err := c.Run(); err != nil {
		return fmt.Errorf("wrangler secret list failed: %w", err)
	}
	return nil
}

func runSecretsSet(key string, args []string, env string) error {
	if _, err := os.Stat("aerostack.toml"); os.IsNotExist(err) {
		return fmt.Errorf("aerostack.toml not found. Run 'aerostack init' first")
	}
	if err := ensureWranglerToml(); err != nil {
		return err
	}

	var value string
	if len(args) >= 2 {
		value = args[1]
	} else {
		// Read from stdin
		var buf [4096]byte
		n, err := os.Stdin.Read(buf[:])
		if err != nil {
			return fmt.Errorf("failed to read value from stdin: %w", err)
		}
		value = strings.TrimSpace(string(buf[:n]))
		if value == "" {
			return fmt.Errorf("value required: pass as argument or pipe via stdin")
		}
	}

	argsWrangler := []string{"-y", "wrangler@latest", "secret", "put", key, "--config", "wrangler.toml"}
	if env != "" {
		argsWrangler = append(argsWrangler, "--env", env)
	}

	c := exec.Command("npx", argsWrangler...)
	c.Dir = "."
	c.Stdin = strings.NewReader(value)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Env = append(os.Environ(), "NPX_UPDATE_NOTIFIER=false")

	absDir, _ := filepath.Abs(".")
	c.Dir = absDir

	if err := c.Run(); err != nil {
		return fmt.Errorf("wrangler secret put failed: %w", err)
	}
	fmt.Printf("✓ Set %s for env %s\n", key, orDefault(env, "production"))
	return nil
}

func orDefault(s, d string) string {
	if s != "" {
		return s
	}
	return d
}
