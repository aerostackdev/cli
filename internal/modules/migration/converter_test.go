package migration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aerostackdev/cli/internal/devserver"
)

func TestConvertWranglerToAerostack(t *testing.T) {
	w := &WranglerConfig{
		Name:              "my-worker",
		Main:              "src/index.ts",
		CompatibilityDate: "2024-01-01",
		D1Databases: []WranglerD1{
			{Binding: "DB", DatabaseName: "my-db", DatabaseID: "abc-123"},
		},
		KVNamespaces: []WranglerKV{
			{Binding: "CACHE", ID: "kv-456"},
		},
	}

	cfg := ConvertWranglerToAerostack(w)

	if cfg.Name != "my-worker" {
		t.Errorf("Name: got %q", cfg.Name)
	}
	if cfg.Main != "src/index.ts" {
		t.Errorf("Main: got %q", cfg.Main)
	}
	if cfg.CompatibilityDate != "2024-01-01" {
		t.Errorf("CompatibilityDate: got %q", cfg.CompatibilityDate)
	}
	if len(cfg.D1Databases) != 1 {
		t.Fatalf("D1Databases: got %d", len(cfg.D1Databases))
	}
	if cfg.D1Databases[0].Binding != "DB" || cfg.D1Databases[0].DatabaseID != "abc-123" {
		t.Errorf("D1Databases[0]: got %+v", cfg.D1Databases[0])
	}
	if len(cfg.KVNamespaces) != 1 {
		t.Fatalf("KVNamespaces: got %d", len(cfg.KVNamespaces))
	}
	if cfg.KVNamespaces[0].Binding != "CACHE" || cfg.KVNamespaces[0].ID != "kv-456" {
		t.Errorf("KVNamespaces[0]: got %+v", cfg.KVNamespaces[0])
	}
}

func TestConvertWranglerToAerostack_EmptyCompatibilityDate(t *testing.T) {
	w := &WranglerConfig{Name: "x", Main: "src/index.ts"}
	cfg := ConvertWranglerToAerostack(w)
	if cfg.CompatibilityDate != "2024-01-01" {
		t.Errorf("default CompatibilityDate: got %q", cfg.CompatibilityDate)
	}
}

func TestGenerateAerostackToml(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "aerostack.toml")

	cfg := &devserver.AerostackConfig{
		Name:              "test",
		Main:              "src/index.ts",
		CompatibilityDate: "2024-01-01",
		D1Databases: []devserver.D1Database{
			{Binding: "DB", DatabaseName: "db", DatabaseID: "id"},
		},
	}

	if err := GenerateAerostackToml(cfg, path); err != nil {
		t.Fatalf("GenerateAerostackToml: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "name = \"test\"") {
		t.Errorf("missing name: %s", content)
	}
	if !strings.Contains(content, "[[d1_databases]]") {
		t.Errorf("missing d1_databases: %s", content)
	}
	if !strings.Contains(content, "binding = \"DB\"") {
		t.Errorf("missing binding: %s", content)
	}
}
