package selfheal

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/aerostackdev/cli/internal/agent"
	"github.com/aerostackdev/cli/internal/api"
	"github.com/aerostackdev/cli/internal/pkg"
)

// ErrorCategory classifies errors for targeted healing
type ErrorCategory string

const (
	ErrorInfrastructure ErrorCategory = "infrastructure"
	ErrorCode           ErrorCategory = "code"
	ErrorAuth           ErrorCategory = "auth"
	ErrorUnknown        ErrorCategory = "unknown"
)

// ClassifyError returns the category of the error for targeted healing
func ClassifyError(err error) ErrorCategory {
	if err == nil {
		return ErrorUnknown
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "connection refused") || strings.Contains(msg, "docker") ||
		strings.Contains(msg, "network") || strings.Contains(msg, "econnrefused") ||
		strings.Contains(msg, "econnreset") {
		return ErrorInfrastructure
	}
	if strings.Contains(msg, "401") || strings.Contains(msg, "unauthorized") ||
		strings.Contains(msg, "invalid token") || strings.Contains(msg, "authentication") ||
		strings.Contains(msg, "api key") || strings.Contains(msg, "forbidden") {
		return ErrorAuth
	}
	if strings.Contains(msg, "syntax") || strings.Contains(msg, "undefined") ||
		strings.Contains(msg, "type error") || strings.Contains(msg, "cannot find") ||
		strings.Contains(msg, "module not found") || strings.Contains(msg, "is not a function") {
		return ErrorCode
	}
	return ErrorUnknown
}

const maxRecursions = 3

type Healer struct {
	agent  *agent.Agent
	logger *pkg.Logger
}

func NewHealer(agent *agent.Agent) *Healer {
	logger, _ := pkg.NewLogger(".") // Default to current dir
	return &Healer{agent: agent, logger: logger}
}

// Start initiates the self-healing loop. Pass fullArgs (e.g. os.Args) to re-run the failed command.
func (h *Healer) Start(ctx context.Context, fullArgs []string, originalErr error) error {
	return h.startRecursive(ctx, fullArgs, originalErr, 0)
}

func (h *Healer) startRecursive(ctx context.Context, fullArgs []string, originalErr error, recursion int) error {
	cmdName := "aerostack"
	if len(fullArgs) > 0 {
		cmdName = fullArgs[0]
		if len(fullArgs) > 1 {
			cmdName = cmdName + " " + strings.Join(fullArgs[1:], " ")
		}
	}

	cat := ClassifyError(originalErr)
	fmt.Printf("\n🚨 Command failed: %v\n", originalErr)
	fmt.Printf("   (classified as: %s)\n", cat)
	fmt.Println("🩹 Aerostack AI is analyzing the error...")

	platformContext := `Aerostack Platform Context:
- Admin Dashboard: https://admin.aerocall.ai
- CLI Keys Settings: https://admin.aerocall.ai/settings
- Account Keys (ak_...) are for CLI deploy/login.
- Project Keys (pk_...) are for database/runtime access.
- Configuration: aerostack.toml in project root.`

	diagnosisPrompt := fmt.Sprintf(`%s

The command failed with error: "%v".
Error category: %s (infrastructure=network/docker, code=syntax/module, auth=401/api key).
Analyze the project context and this error.
Explain what went wrong and propose a fix.
If it's an Auth error, suggest where the user can find the correct key in the Admin UI.
If the fix involves editing a file, use the 'write_file' tool with the full corrected file content.
Use read_file and list_dir to gather context first. Be concise.`, platformContext, originalErr, cat)

	if h.logger != nil {
		h.logger.Error(fmt.Sprintf("Command failed: %s", cmdName), originalErr)
	}

	proposal, edits, err := h.agent.ResolveForHealing(ctx, diagnosisPrompt)
	if err != nil {
		return fmt.Errorf("AI diagnosis failed: %w", err)
	}

	if proposal == "" {
		proposal = "No specific proposal. Check the error above."
	}

	diff := buildDiff(edits)

	accepted, tuiErr := ShowProposal(originalErr, proposal, diff)
	if tuiErr != nil {
		return fmt.Errorf("TUI error: %w", tuiErr)
	}

	if !accepted {
		fmt.Println("\nFix not applied. Exiting.")
		if h.logger != nil {
			logs, _ := h.logger.GetLogContent()
			apiKey := pkg.GetApiKeyFromToml()
			if apiKey != "" {
				api.SendTelemetry(apiKey, "cli-healing-rejected", logs)
			}
		}
		return nil
	}

	// Apply edits
	for _, e := range edits {
		if err := os.MkdirAll(filepath.Dir(e.Path), 0755); err != nil {
			fmt.Printf("⚠️  Failed to create directory for %s: %v\n", e.Path, err)
			continue
		}
		if err := os.WriteFile(e.Path, []byte(e.Content), 0644); err != nil {
			fmt.Printf("⚠️  Failed to write %s: %v\n", e.Path, err)
		} else {
			fmt.Printf("✅ Applied fix to %s\n", e.Path)
		}
	}

	// Re-run the command
	if len(fullArgs) < 1 {
		fmt.Println("Cannot re-run: no command args.")
		return nil
	}

	fmt.Println("\n🔄 Re-running command...")
	cmd := exec.Command(fullArgs[0], fullArgs[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Dir, _ = os.Getwd()

	if err := cmd.Run(); err != nil {
		if recursion >= maxRecursions-1 {
			fmt.Printf("\n❌ Command still failing after %d attempts: %v\n", maxRecursions, err)
			// Send telemetry on failure
			if h.logger != nil {
				logs, _ := h.logger.GetLogContent()
				apiKey := pkg.GetApiKeyFromToml()
				if apiKey != "" {
					api.SendTelemetry(apiKey, "cli-healing-failed", logs)
				}
			}
			return err
		}
		fmt.Printf("\n⚠️  Command failed again with a different error. Retrying (attempt %d/%d)...\n", recursion+2, maxRecursions)
		return h.startRecursive(ctx, fullArgs, err, recursion+1)
	}

	fmt.Println("\n✅ Command succeeded after fix!")

	// Send telemetry on success
	if h.logger != nil {
		logs, _ := h.logger.GetLogContent()
		apiKey := pkg.GetApiKeyFromToml()
		if apiKey != "" {
			api.SendTelemetry(apiKey, "cli-healing", logs)
		}
	}

	return nil
}

func buildDiff(edits []agent.FileEdit) string {
	if len(edits) == 0 {
		return ""
	}
	var b strings.Builder
	for _, e := range edits {
		b.WriteString(fmt.Sprintf("--- %s ---\n", e.Path))
		lines := strings.Split(e.Content, "\n")
		for i, line := range lines {
			if i < 20 {
				b.WriteString("+ " + line + "\n")
			}
		}
		if len(lines) > 20 {
			b.WriteString(fmt.Sprintf("... (%d more lines)\n", len(lines)-20))
		}
		b.WriteString("\n")
	}
	return b.String()
}
