package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

// NewAddCommand creates the 'aerostack add' command with function and lib subcommands
func NewAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a function or shared lib to your project",
		Long: `Add new services or shared code to your Aerostack project.

  aerostack add function <name>  — Create a new Worker service
  aerostack add lib <name>       — Create a shared module in shared/`,
	}

	cmd.AddCommand(NewAddFunctionCommand())
	cmd.AddCommand(NewAddLibCommand())
	return cmd
}

// NewAddFunctionCommand creates 'aerostack add function <name>'
func NewAddFunctionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "function [name]",
		Short: "Add a new Worker function/service",
		Long: `Create a new Worker service at services/<name>/index.ts.
Registers the service in aerostack.toml for multi-worker projects.

Example:
  aerostack add function api-gateway`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return addFunction(args[0])
		},
	}
}

// NewAddLibCommand creates 'aerostack add lib <name>'
func NewAddLibCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "lib [name]",
		Short: "Add a shared library module",
		Long: `Create a new shared module at shared/<name>.ts.
Import via: import { ... } from "@shared/<name>"

Example:
  aerostack add lib auth`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return addLib(args[0])
		},
	}
}

func addFunction(name string) error {
	// Validate name (alphanumeric, hyphens)
	if !isValidName(name) {
		return fmt.Errorf("invalid name %q: use alphanumeric and hyphens only (e.g. api-gateway)", name)
	}

	if _, err := os.Stat("aerostack.toml"); os.IsNotExist(err) {
		return fmt.Errorf("aerostack.toml not found. Run 'aerostack init' first")
	}

	serviceDir := filepath.Join("services", name)
	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		return fmt.Errorf("failed to create services/%s: %w", name, err)
	}

	indexPath := filepath.Join(serviceDir, "index.ts")
	content := fmt.Sprintf(`// Service: %s
// Import shared code: import { getDb } from "@shared/db"

export default {
  async fetch(request: Request, env: Record<string, unknown>, ctx: ExecutionContext): Promise<Response> {
    return new Response("Hello from %s!");
  },
};
`, name, name)

	if err := os.WriteFile(indexPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", indexPath, err)
	}

	// Append [[services]] to aerostack.toml if not already present
	if err := appendServiceToAerostackToml(name, filepath.Join("services", name, "index.ts")); err != nil {
		return err
	}

	fmt.Printf("✅ Created %s\n", indexPath)
	fmt.Printf("   Registered in aerostack.toml (multi-worker support coming soon)\n")
	return nil
}

func addLib(name string) error {
	// Validate name (alphanumeric, hyphens)
	if !isValidName(name) {
		return fmt.Errorf("invalid name %q: use alphanumeric and hyphens only (e.g. auth)", name)
	}

	if _, err := os.Stat("aerostack.toml"); os.IsNotExist(err) {
		return fmt.Errorf("aerostack.toml not found. Run 'aerostack init' first")
	}

	sharedDir := "shared"
	if err := os.MkdirAll(sharedDir, 0755); err != nil {
		return fmt.Errorf("failed to create shared/: %w", err)
	}

	libPath := filepath.Join(sharedDir, name+".ts")
	content := fmt.Sprintf(`// Shared module: %s
// Import via: import { ... } from "@shared/%s"

export function placeholder(): string {
  return "TODO: implement %s";
}
`, name, name, name)

	if _, err := os.Stat(libPath); err == nil {
		return fmt.Errorf("%s already exists", libPath)
	}

	if err := os.WriteFile(libPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", libPath, err)
	}

	fmt.Printf("✅ Created %s\n", libPath)
	fmt.Printf("   Import via: import { ... } from \"@shared/%s\"\n", name)
	return nil
}

func isValidName(s string) bool {
	matched, _ := regexp.MatchString(`^[a-z0-9][a-z0-9-]*[a-z0-9]$|^[a-z0-9]$`, strings.ToLower(s))
	return matched
}

func appendServiceToAerostackToml(serviceName, mainPath string) error {
	data, err := os.ReadFile("aerostack.toml")
	if err != nil {
		return err
	}
	content := string(data)

	// Check if [[services]] block for this service already exists
	blockPattern := regexp.MustCompile(`\[\[services\]\]\s*\n[^\[]*name\s*=\s*"` + regexp.QuoteMeta(serviceName) + `"`)
	if blockPattern.MatchString(content) {
		return nil // Already registered
	}

	// Append new [[services]] block
	block := fmt.Sprintf("\n# Multi-worker: service %s\n[[services]]\nname = %q\nmain = %q\n", serviceName, serviceName, mainPath)
	if err := os.WriteFile("aerostack.toml", append(data, []byte(block)...), 0644); err != nil {
		return fmt.Errorf("failed to update aerostack.toml: %w", err)
	}
	return nil
}
