package migration

import (
	"context"
	"fmt"

	"github.com/aerostackdev/cli/internal/agent"
)

type RefactorAI struct {
	agent *agent.Agent
}

func NewRefactorAI(agent *agent.Agent) *RefactorAI {
	return &RefactorAI{agent: agent}
}

func (r *RefactorAI) CheckCompatibility(ctx context.Context, entryFile string) error {
	fmt.Printf("🕵️  Checking code compatibility for %s...\n", entryFile)

	prompt := fmt.Sprintf(`You are the Aerostack Migration Assistant.
We are migrating a Cloudflare Worker project to the Aerostack standard.
Target file: "%s"

Task:
1. proper check the code for environmental variable usage (process.env vs env object).
2. Check for 'wrangler' specific bindings that might need updates.
3. If the code looks compatible, say "Compatible".
4. If changes are needed, use the 'write_file' tool to update the code to be compatible with standard ES Module Worker syntax (export default { fetch(request, env, ctx) ... }).

Current context is a standard Cloudflare Worker.`, entryFile)

	return r.agent.Resolve(ctx, prompt)
}
