package pkg

import (
	"fmt"
	"os"
	"regexp"
)

// AppendServiceToAerostackToml adds a service entry to the aerostack.toml file if it doesn't already exist.
func AppendServiceToAerostackToml(serviceName, mainPath string) error {
	data, err := os.ReadFile("aerostack.toml")
	if err != nil {
		return err
	}
	content := string(data)

	// Check if [[services]] block for this service already exists
	// Matching [[services]] followed by name = "serviceName"
	blockPattern := regexp.MustCompile(`(?s)\[\[services\]\].*?name\s*=\s*"` + regexp.QuoteMeta(serviceName) + `"`)
	if blockPattern.MatchString(content) {
		return nil // Already registered
	}

	// Append new [[services]] block
	block := fmt.Sprintf("\n# Community: service %s\n[[services]]\nname = %q\nmain = %q\n", serviceName, serviceName, mainPath)
	if err := os.WriteFile("aerostack.toml", append(data, []byte(block)...), 0644); err != nil {
		return fmt.Errorf("failed to update aerostack.toml: %w", err)
	}
	return nil
}

// GetApiKeyFromToml reads the api_key from aerostack.toml if it exists.
func GetApiKeyFromToml() string {
	data, err := os.ReadFile("aerostack.toml")
	if err != nil {
		return ""
	}

	// We can use a simple regex or a full TOML parser.
	// Since we already have go-toml/v2 in go.mod, let's use it for robustness.
	// But to avoid a heavy dependency in this simple utility, let's try regex first
	// as it's often used for quick lookups in this CLI.

	// Match api_key = "..." or api_key="..." in [env.production] or root
	re := regexp.MustCompile(`(?m)^\s*api_key\s*=\s*"([^"]+)"`)
	matches := re.FindStringSubmatch(string(data))
	if len(matches) > 1 {
		return matches[1]
	}

	return ""
}
