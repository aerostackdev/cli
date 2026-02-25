package devserver

import (
	"os"
	"strings"
	"testing"
)

func TestGenerateWranglerToml(t *testing.T) {
	cfg := &AerostackConfig{
		Name:              "test-app",
		CompatibilityDate: "2024-01-01",
		Vars: map[string]string{
			"CUSTOM_VAR": "custom-value",
		},
		Services: []Service{
			{Name: "auth", Main: "src/auth.ts"},
		},
	}

	outputPath := "test-wrangler.toml"
	defer os.Remove(outputPath)

	err := GenerateWranglerToml(cfg, outputPath)
	if err != nil {
		t.Fatalf("GenerateWranglerToml failed: %v", err)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read generated toml: %v", err)
	}
	content := string(data)

	// Check for AEROSTACK_API_URL injection
	if !strings.Contains(content, `AEROSTACK_API_URL = "http://localhost:8787"`) {
		t.Errorf("AEROSTACK_API_URL not found in generated toml")
	}

	// Check for custom vars
	if !strings.Contains(content, `CUSTOM_VAR = "custom-value"`) {
		t.Errorf("CUSTOM_VAR not found in generated toml")
	}

	// Check for service bindings
	if !strings.Contains(content, `[[services]]`) {
		t.Errorf("[[services]] block not found in generated toml")
	}
	if !strings.Contains(content, `binding = "AUTH"`) {
		t.Errorf("AUTH binding not found in generated toml")
	}
	if !strings.Contains(content, `service = "test-app-auth"`) {
		t.Errorf("auth service name mismatch in generated toml")
	}
}

func TestParseAerostackToml(t *testing.T) {
	content := `
name = "test-app"
project_slug = "test-slug"

[vars]
API_URL = "https://custom-api.com"
OTHER_VAR = "other-value"
`
	tmpfile, err := os.CreateTemp("", "aerostack-*.toml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	cfg, err := ParseAerostackToml(tmpfile.Name())
	if err != nil {
		t.Fatalf("ParseAerostackToml failed: %v", err)
	}

	if cfg.Vars["API_URL"] != "https://custom-api.com" {
		t.Errorf("expected API_URL to be https://custom-api.com, got %s", cfg.Vars["API_URL"])
	}
	if cfg.Vars["OTHER_VAR"] != "other-value" {
		t.Errorf("expected OTHER_VAR to be other-value, got %s", cfg.Vars["OTHER_VAR"])
	}
}
