package auth

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aerostackdev/cli/internal/agent"
)

type Doctor struct {
	agent *agent.Agent
}

func NewDoctor(agent *agent.Agent) *Doctor {
	return &Doctor{agent: agent}
}

// Run scans the project for authentication issues
func (d *Doctor) Run(ctx context.Context) error {
	fmt.Println("🩺 Running Auth Doctor...")

	// 1. Gather Context
	// We'll look for specific files that are critical for Auth
	filesToCheck := []string{".env", "packages/api/wrangler.toml", "packages/api/src/auth-hooks.ts"}

	var contextBuilder strings.Builder
	contextBuilder.WriteString("Project Auth Configuration Analysis:\n")

	issuesFound := 0

	for _, relPath := range filesToCheck {
		cwd, _ := os.Getwd()
		absPath := filepath.Join(cwd, relPath)

		content, err := os.ReadFile(absPath)
		if err != nil {
			if os.IsNotExist(err) {
				contextBuilder.WriteString(fmt.Sprintf("- [MISSING] %s\n", relPath))
				issuesFound++
			} else {
				contextBuilder.WriteString(fmt.Sprintf("- [ERROR] Could not read %s: %v\n", relPath, err))
			}
			continue
		}

		contextBuilder.WriteString(fmt.Sprintf("- [OK] %s found. Content preview:\n%s\n---\n", relPath, string(content)))
	}

	if issuesFound == 0 {
		fmt.Println("✅ Key modules found.")
	} else {
		fmt.Printf("⚠️  %d issues potentially found with file layout.\n", issuesFound)
	}

	// 2. AI Diagnosis
	fmt.Println("🧠 Analyzing configuration with Aerostack AI...")

	prompt := fmt.Sprintf(`You are the Aerostack Auth Doctor.
Review the following project context for authentication configuration.
Check for:
1. Missing AUTH_SECRET in .env or wrangler.toml
2. CORS origin mismatches (if evident)
3. Invalid JWT/Session expiry settings
4. Any potential security risks

Context:
%s

Report any issues found and suggest exact fixes. If everything looks good, say "Everything looks correct."`, contextBuilder.String())

	return d.agent.Resolve(ctx, prompt)
}
