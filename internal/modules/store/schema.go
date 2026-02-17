package store

import (
	"context"
	"fmt"

	"github.com/aerostackdev/cli/internal/agent"
)

type SchemaAI struct {
	agent *agent.Agent
}

func NewSchemaAI(agent *agent.Agent) *SchemaAI {
	return &SchemaAI{agent: agent}
}

// Generate proposes schema changes based on user intent
func (s *SchemaAI) Generate(ctx context.Context, intent string) error {
	fmt.Printf("🏗️  Generating schema for: '%s'...\n", intent)

	// 1. Read existing schema (if any)
	// In a real app, we'd read packages/api/prisma/schema.prisma or migrations
	// For D1, we might look at existing SQL files

	schemaContext := "Current Schema: (Unknown/Empty)"
	// Placeholder: In real impl, read 'packages/api/migrations'

	prompt := fmt.Sprintf(`You are the Aerostack Store Architect.
User Intent: "%s"

%s

Task:
1. Design the SQL schema changes needed for this feature (SQLite compatible).
2. implementation should be in a single .sql file content.
3. Suggest a filename for the migration (e.g., 0001_add_users.sql).
4. Use the 'write_file' tool to save the migration to 'packages/api/migrations/<filename>'.

Refer to the D1 documentation for data types (TEXT, INTEGER, REAL, BLOB).`, intent, schemaContext)

	return s.agent.Resolve(ctx, prompt)
}
