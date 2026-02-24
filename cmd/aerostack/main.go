package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aerostackdev/cli/internal/agent"
	"github.com/aerostackdev/cli/internal/api"
	"github.com/aerostackdev/cli/internal/commands"
	"github.com/aerostackdev/cli/internal/credentials"
	"github.com/aerostackdev/cli/internal/pkg"
	"github.com/aerostackdev/cli/internal/selfheal"
	"github.com/spf13/cobra"
)

var (
	version = "v1.5.3"
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
		Version:       fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date),
		SilenceErrors: true, // We handle errors manually
		SilenceUsage:  true,
	}

	// Add subcommands
	rootCmd.AddCommand(commands.NewInitCommand())
	rootCmd.AddCommand(commands.NewDevCommand())
	rootCmd.AddCommand(commands.NewDeployCommand())
	rootCmd.AddCommand(commands.NewLoginCommand())
	rootCmd.AddCommand(commands.NewLinkCommand())
	rootCmd.AddCommand(commands.NewWhoamiCommand())
	rootCmd.AddCommand(commands.NewDBCommand())
	rootCmd.AddCommand(commands.NewResourcesCommand())
	rootCmd.AddCommand(commands.NewGenerateCommand())
	rootCmd.AddCommand(commands.NewAddCommand())
	rootCmd.AddCommand(commands.NewTestCommand())
	rootCmd.AddCommand(commands.NewSecretsCommand())
	rootCmd.AddCommand(commands.NewIndexCommand())
	rootCmd.AddCommand(commands.NewAICommand())
	rootCmd.AddCommand(commands.NewAuthCommand())
	rootCmd.AddCommand(commands.NewStoreCommand())
	rootCmd.AddCommand(commands.NewUICommand())
	rootCmd.AddCommand(commands.NewFunctionsCommand())
	rootCmd.AddCommand(commands.NewMigrateCommand())

	if err := rootCmd.Execute(); err != nil {
		// 1. Basic error print
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)

		// 2. Global Telemetry: Push all logs to server for visibility
		apiKey := credentials.GetAPIKey()
		if apiKey != "" {
			// Get current logs if any
			cwd, _ := os.Getwd()
			logger, _ := pkg.NewLogger(cwd)
			var logs string
			if logger != nil {
				logs, _ = logger.GetLogContent()
			}
			if logs == "" {
				logs = fmt.Sprintf("Command failed: %s\nError: %v", strings.Join(os.Args, " "), err)
			}
			// Use "cli-error" as projectID placeholder if not in a project
			api.SendTelemetry(apiKey, "cli-error", logs)
		}

		// 3. Check if we should trigger Self-Healing
		// For MVP, trigger on ANY error if OPENAI_API_KEY is present
		if shouldHeal(err) {
			ctx := context.Background()
			cwd, _ := os.Getwd()

			// Init PKG & Agent (lite version, no error if missing)
			store, _ := pkg.NewStore(cwd)
			if store != nil {
				ag, agentErr := agent.NewAgent(store, false)
				if agentErr == nil {
					healer := selfheal.NewHealer(ag)
					if healErr := healer.Start(ctx, os.Args, err); healErr != nil {
						fmt.Fprintf(os.Stderr, "Self-healing failed: %v\n", healErr)
					}
				}
			}
		}

		os.Exit(1)
	}
}

func shouldHeal(err error) bool {
	// Don't heal on simple usage errors or interruptions
	msg := err.Error()
	if strings.Contains(msg, "interrupt") || strings.Contains(msg, "canceled") {
		return false
	}
	// Check for API key: Azure, OpenAI, Anthropic, or Aerostack backend (aerostack.toml or credentials)
	if os.Getenv("AZURE_OPENAI_API_KEY") != "" || os.Getenv("OPENAI_API_KEY") != "" || os.Getenv("ANTHROPIC_API_KEY") != "" {
		return true
	}
	if credentials.GetAPIKey() != "" {
		return true
	}
	return false
}
