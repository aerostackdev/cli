package store

import (
	"context"
	"fmt"

	"github.com/aerostackdev/cli/internal/agent"
)

type SeederAI struct {
	agent *agent.Agent
}

func NewSeederAI(agent *agent.Agent) *SeederAI {
	return &SeederAI{agent: agent}
}

func (s *SeederAI) Seed(ctx context.Context, smart bool) error {
	if !smart {
		fmt.Println("🌱 Running standard seed...")
		// Call standard seed script (e.g., npm run seed)
		return nil
	}

	fmt.Println("🧠 Running Smart Seed...")

	// 1. Analyze Schema to understand what to seed
	prompt := `You are the Aerostack Smart Seeder.
Analyze the project's SQL migrations in 'packages/api/migrations'.
Understand the data model (Products, Users, etc.).
Generate a SQL script to insert 5-10 realistic demo records for the main tables.
Use the 'write_file' tool to save this to 'packages/api/seed_custom.sql'.
Then, instruct the user to run 'npx wrangler d1 execute ...' to apply it.`

	return s.agent.Resolve(ctx, prompt)
}
