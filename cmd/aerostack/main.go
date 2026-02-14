package main

import (
	"fmt"
	"os"

	"github.com/aerostackdev/cli/internal/commands"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "aerostack",
		Short: "Aerostack CLI - Build and deploy serverless applications with ease",
		Long: `Aerostack CLI is a powerful command-line tool for building, testing,
and deploying serverless applications on Cloudflare's edge infrastructure.

Features:
  • Zero-config local development environment
  • Multi-database orchestration (D1, Neon, External)
  • Built-in testing and deployment workflows
  • AI-powered error fixing and code generation
  • Production-ready starter templates`,
		Version: fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date),
	}

	// Add subcommands
	rootCmd.AddCommand(commands.NewInitCommand())
	rootCmd.AddCommand(commands.NewDevCommand())
	rootCmd.AddCommand(commands.NewDeployCommand())
	rootCmd.AddCommand(commands.NewLoginCommand())
	rootCmd.AddCommand(commands.NewDBCommand())
	rootCmd.AddCommand(commands.NewGenerateCommand())
	rootCmd.AddCommand(commands.NewAddCommand())
	rootCmd.AddCommand(commands.NewTestCommand())
	rootCmd.AddCommand(commands.NewSecretsCommand())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
