// Package provision auto-provisions Cloudflare resources (D1, KV, R2, Queues, etc.)
// in the user's account when deploying with --cloudflare.
// Resources are created based on aerostack.toml config; placeholders are replaced with real IDs.
package provision

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/aerostackdev/cli/internal/devserver"
)

// UUID regex for parsing wrangler output
var uuidRe = regexp.MustCompile(`[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}`)

// Placeholder patterns that indicate we need to create the resource
func isPlaceholderID(id string) bool {
	if id == "" {
		return true
	}
	id = strings.TrimSpace(id)
	// Common placeholders
	if strings.HasPrefix(id, "YOUR_") || strings.HasPrefix(id, "your_") {
		return true
	}
	if id == "local-mock" || id == "aerostack-local" || id == "local" {
		return true
	}
	// Not a valid UUID format = placeholder
	if !uuidRe.MatchString(id) {
		return true
	}
	return false
}

// ProvisionCloudflareResources creates D1, KV, R2, Queues (etc.) in the user's
// Cloudflare account when IDs are placeholders. Updates aerostack.toml with real IDs.
// Extensible: add more resource types (Vectorize, AI, etc.) as SDK supports them.
func ProvisionCloudflareResources(cfg *devserver.AerostackConfig, env string, projectRoot string) error {
	// Get D1 config for this env
	dbs := cfg.EnvOverrides[env]
	if len(dbs) == 0 {
		dbs = cfg.D1Databases
	}

	modified := false
	for i := range dbs {
		db := &dbs[i]
		if !isPlaceholderID(db.DatabaseID) {
			continue
		}

		fmt.Printf("   D1 (%s): creating %q in your account...\n", db.Binding, db.DatabaseName)
		id, err := createD1(db.DatabaseName, projectRoot)
		if err != nil {
			return fmt.Errorf("D1 create failed for %s: %w", db.DatabaseName, err)
		}

		oldID := db.DatabaseID
		if oldID == "" {
			oldID = "local-mock"
		}
		db.DatabaseID = id
		fmt.Printf("   ✓ Created D1 %q → %s\n", db.DatabaseName, id)

		// Update aerostack.toml (db.DatabaseID already updated in-memory)
		configPath := filepath.Join(projectRoot, "aerostack.toml")
		if err := updateConfigDatabaseID(configPath, oldID, id, env, db.DatabaseName); err != nil {
			fmt.Printf("   ⚠ Could not update aerostack.toml: %v (ID saved for this run)\n", err)
		} else {
			modified = true
		}
	}

	if modified {
		fmt.Println("   ✓ Updated aerostack.toml")
	}

	return nil
}

func createD1(databaseName string, projectRoot string) (string, error) {
	cmd := exec.Command("npx", "-y", "wrangler@latest", "d1", "create", databaseName)
	cmd.Dir = projectRoot
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("wrangler d1 create: %w", err)
	}

	// Parse UUID from output (wrangler prints database_id = "uuid" or similar)
	matches := uuidRe.FindStringSubmatch(string(out))
	if len(matches) == 0 {
		return "", fmt.Errorf("could not parse database ID from wrangler output")
	}
	return matches[0], nil
}

// updateConfigDatabaseID replaces a placeholder database_id in aerostack.toml.
func updateConfigDatabaseID(path, oldID, newID, env, databaseName string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	content := string(data)
	oldPattern := `database_id = "` + regexp.QuoteMeta(oldID) + `"`
	newReplacement := `database_id = "` + newID + `"`

	// Replace first occurrence (covers base [[d1_databases]] or [[env.X.d1_databases]])
	content = strings.Replace(content, oldPattern, newReplacement, 1)

	return os.WriteFile(path, []byte(content), 0644)
}
