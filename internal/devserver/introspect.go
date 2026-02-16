package devserver

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os/exec"
	"sort"
	"strings"

	"github.com/aerostackdev/cli/internal/api"
	_ "github.com/lib/pq"
)

// WranglerResult represents the output of wrangler d1 execute --json
type WranglerResult struct {
	Results []map[string]interface{} `json:"results"`
	Success bool                     `json:"success"`
}

// TableSchema represents the structure of a database table
// SourceBinding: D1/Postgres binding name for namespacing (e.g. "DB", "PgDb")
type TableSchema struct {
	Name          string
	Columns       []ColumnSchema
	SourceBinding string // e.g. "DB" or "PgDb" — used to avoid collisions across databases
}

// ColumnSchema represents a column in a table
type ColumnSchema struct {
	Name       string
	Type       string
	IsNullable bool
	IsPrimary  bool
}

// IntrospectD1Local introspects a local D1 database via Wrangler.
// projectRoot: directory containing wrangler.toml (where wrangler must run from).
// sourceBinding: binding name for namespacing (e.g. "DB").
func IntrospectD1Local(dbName, projectRoot, sourceBinding string) ([]TableSchema, error) {
	// Use wrangler d1 execute --json to get table list
	cmd := exec.Command("npx", "wrangler", "d1", "execute", dbName, "--local", "--command", "SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' AND name NOT LIKE '_cf_%'", "--json")
	cmd.Dir = projectRoot
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list D1 tables: %w (output: %s)", err, string(out))
	}

	var results []WranglerResult
	if err := json.Unmarshal(out, &results); err != nil {
		return nil, fmt.Errorf("failed to parse D1 tables JSON: %w (output: %s)", err, string(out))
	}

	if len(results) == 0 || !results[0].Success {
		return nil, fmt.Errorf("D1 introspection failed or returned no results")
	}

	var schemas []TableSchema
	for _, row := range results[0].Results {
		name, ok := row["name"].(string)
		if !ok || name == "" {
			continue
		}

		colCmd := exec.Command("npx", "wrangler", "d1", "execute", dbName, "--local", "--command", fmt.Sprintf("PRAGMA table_info(%s)", name), "--json")
		colCmd.Dir = projectRoot
		colOut, err := colCmd.Output()
		if err != nil {
			continue
		}

		var colResults []WranglerResult
		if err := json.Unmarshal(colOut, &colResults); err != nil {
			continue
		}

		if len(colResults) == 0 || !colResults[0].Success {
			continue
		}

		var cols []ColumnSchema
		for _, colRow := range colResults[0].Results {
			colName, _ := colRow["name"].(string)
			colType, _ := colRow["type"].(string)
			notNull, _ := colRow["notnull"].(float64)
			pk, _ := colRow["pk"].(float64)

			cols = append(cols, ColumnSchema{
				Name:       colName,
				Type:       mapSQLiteType(colType),
				IsNullable: notNull == 0,
				IsPrimary:  pk == 1,
			})
		}

		schemas = append(schemas, TableSchema{
			Name:          name,
			Columns:       cols,
			SourceBinding: sourceBinding,
		})
	}

	return schemas, nil
}

// IntrospectPostgres introspects an external Postgres database.
// sourceBinding: binding name for namespacing (e.g. "PgDb").
func IntrospectPostgres(connStr, sourceBinding string) ([]TableSchema, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Postgres: %w", err)
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_type = 'BASE TABLE'
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to list Postgres tables: %w", err)
	}
	defer rows.Close()

	var schemas []TableSchema
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			continue
		}

		colRows, err := db.Query(`
			SELECT column_name, data_type, is_nullable, 
			       EXISTS (
				       SELECT 1 FROM information_schema.key_column_usage kcu
				       JOIN information_schema.table_constraints tc ON kcu.constraint_name = tc.constraint_name
				       WHERE kcu.table_name = $1 AND kcu.column_name = information_schema.columns.column_name 
				       AND tc.constraint_type = 'PRIMARY KEY'
			       ) as is_primary
			FROM information_schema.columns
			WHERE table_name = $1
		`, name)
		if err != nil {
			continue
		}

		var cols []ColumnSchema
		for colRows.Next() {
			var colName, dataType, isNullable string
			var isPrimary bool
			if err := colRows.Scan(&colName, &dataType, &isPrimary, &isNullable); err != nil {
				continue
			}
			cols = append(cols, ColumnSchema{
				Name:       colName,
				Type:       mapPostgresType(dataType),
				IsNullable: isNullable == "YES",
				IsPrimary:  isPrimary,
			})
		}
		colRows.Close()

		schemas = append(schemas, TableSchema{
			Name:          name,
			Columns:       cols,
			SourceBinding: sourceBinding,
		})
	}

	return schemas, nil
}

// GenerateTypeScript produces the TypeScript interfaces for the schemas.
// Tables from different databases are namespaced to avoid collisions (e.g. DBUsers vs PgDbUsers).
func GenerateTypeScript(schemas []TableSchema, meta *api.ProjectMetadata) string {
	var sb strings.Builder
	sb.WriteString("// Generated by Aerostack CLI. Do not edit manually.\n\n")

	// 1. Table Interfaces
	sort.Slice(schemas, func(i, j int) bool {
		a, b := schemas[i], schemas[j]
		if a.SourceBinding != b.SourceBinding {
			return a.SourceBinding < b.SourceBinding
		}
		return a.Name < b.Name
	})

	for _, table := range schemas {
		baseName := toPascalCase(table.Name)
		interfaceName := baseName
		if table.SourceBinding != "" {
			interfaceName = toPascalCase(table.SourceBinding) + baseName
		}
		sb.WriteString(fmt.Sprintf("export interface %s {\n", interfaceName))
		for _, col := range table.Columns {
			nullable := ""
			if col.IsNullable {
				nullable = "?"
			}
			sb.WriteString(fmt.Sprintf("  %s%s: %s;\n", col.Name, nullable, col.Type))
		}
		sb.WriteString("}\n\n")
	}

	// 2. Collection Data Types (from metadata)
	if meta != nil {
		for _, col := range meta.Collections {
			interfaceName := toPascalCase(col.Slug) + "Item"
			sb.WriteString(fmt.Sprintf("export interface %s {\n", interfaceName))
			sb.WriteString("  id: string;\n")
			sb.WriteString("  slug: string;\n")
			sb.WriteString("  data: any;\n")
			sb.WriteString("  created_at: number;\n")
			sb.WriteString("}\n\n")
		}
	}

	// 3. Custom API Types
	if meta != nil && len(meta.Hooks) > 0 {
		sb.WriteString("export interface CustomApiSchema {\n")
		for _, hook := range meta.Hooks {
			sb.WriteString(fmt.Sprintf("  %q: {\n", hook.Slug))
			sb.WriteString("    params: any;\n")
			sb.WriteString("    response: any;\n")
			sb.WriteString("  };\n")
		}
		sb.WriteString("}\n\n")
	}

	// 4. Final Project Schema
	sb.WriteString("export interface ProjectSchema {\n")

	// Collections section
	sb.WriteString("  collections: {\n")
	if meta != nil {
		for _, col := range meta.Collections {
			interfaceName := toPascalCase(col.Slug) + "Item"
			sb.WriteString(fmt.Sprintf("    %q: %s;\n", col.Slug, interfaceName))
		}
	}
	sb.WriteString("  };\n")

	// Custom APIs section
	sb.WriteString("  customApis: ")
	if meta != nil && len(meta.Hooks) > 0 {
		sb.WriteString("CustomApiSchema;\n")
	} else {
		sb.WriteString("Record<string, { params: any; response: any }>;\n")
	}

	// Database section
	sb.WriteString("  db: {\n")
	for _, table := range schemas {
		interfaceName := toPascalCase(table.Name)
		if table.SourceBinding != "" {
			interfaceName = toPascalCase(table.SourceBinding) + interfaceName
		}
		key := table.Name
		if table.SourceBinding != "" {
			key = table.SourceBinding + "." + table.Name
		}
		sb.WriteString(fmt.Sprintf("    %q: %s;\n", key, interfaceName))
	}
	sb.WriteString("  };\n")

	sb.WriteString("  queues: Record<string, any>;\n")
	sb.WriteString("  cache: Record<string, any>;\n")
	sb.WriteString("}\n")

	return sb.String()
}

func parseTableNames(out string) []string {
	var names []string
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "┌") || strings.HasPrefix(trimmed, "├") || strings.HasPrefix(trimmed, "└") || strings.Contains(trimmed, "name") {
			continue
		}
		if strings.HasPrefix(trimmed, "│") {
			parts := strings.Split(trimmed, "│")
			if len(parts) > 1 {
				name := strings.TrimSpace(parts[1])
				if name != "" && name != "name" {
					names = append(names, name)
				}
			}
		}
	}
	return names
}

func mapSQLiteType(t string) string {
	t = strings.ToUpper(t)
	switch {
	case strings.Contains(t, "INT"):
		return "number"
	case strings.Contains(t, "TEXT"), strings.Contains(t, "CHAR"), t == "UUID":
		return "string"
	case strings.Contains(t, "REAL"), strings.Contains(t, "FLOAT"), strings.Contains(t, "DOUBLE"):
		return "number"
	case strings.Contains(t, "BLOB"):
		return "Uint8Array"
	case strings.Contains(t, "BOOL"):
		return "boolean"
	default:
		return "any"
	}
}

func mapPostgresType(t string) string {
	t = strings.ToLower(t)
	switch {
	case strings.Contains(t, "int"), t == "numeric", t == "real", t == "double precision":
		return "number"
	case strings.Contains(t, "text"), strings.Contains(t, "char"), t == "uuid":
		return "string"
	case t == "boolean":
		return "boolean"
	case t == "timestamp", t == "date":
		return "string"
	case t == "json", t == "jsonb":
		return "any"
	default:
		return "any"
	}
}

func toPascalCase(s string) string {
	words := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-'
	})
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
		}
	}
	return strings.Join(words, "")
}
