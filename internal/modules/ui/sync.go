package ui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aerostackdev/cli/internal/agent"
)

type UISync struct {
	agent *agent.Agent
}

func NewUISync(agent *agent.Agent) *UISync {
	return &UISync{agent: agent}
}

// Sync scans the project's UI configuration and prepares it for the Agent
func (u *UISync) Sync(ctx context.Context) error {
	fmt.Println("🎨 Syncing UI Theme Context...")

	// 1. Scan for config files
	filesToScan := []string{
		"tailwind.config.ts",
		"tailwind.config.js",
		"apps/web/tailwind.config.ts", // Monorepo support
		"packages/ui/src/globals.css",
	}

	var themeContext strings.Builder
	themeContext.WriteString("Project UI Theme Configuration:\n")
	found := 0

	cwd, _ := os.Getwd()
	for _, relPath := range filesToScan {
		absPath := filepath.Join(cwd, relPath)
		content, err := os.ReadFile(absPath)
		if err == nil {
			themeContext.WriteString(fmt.Sprintf("\n--- %s ---\n%s\n", relPath, string(content)))
			found++
			fmt.Printf("✔ Found %s\n", relPath)
		}
	}

	if found == 0 {
		fmt.Println("⚠️  No UI configuration files found. Agent will use default styling.")
		return nil
	}

	// 2. Interpret with Agent (optional, but good for summarization)
	// For now, we'll just save this "context summary" to a file that the Agent knows to look for
	// In the future, this could be stored in the PKG.

	agentContextPath := filepath.Join(cwd, ".aerostack", "context", "ui_theme.md")
	if err := os.MkdirAll(filepath.Dir(agentContextPath), 0755); err != nil {
		return err
	}

	err := os.WriteFile(agentContextPath, []byte(themeContext.String()), 0644)
	if err != nil {
		return fmt.Errorf("failed to save UI context: %w", err)
	}

	fmt.Printf("✨ UI Context synced to %s\n", agentContextPath)
	fmt.Println("The AI Agent will now be aware of your custom colors, fonts, and spacing.")

	return nil
}
