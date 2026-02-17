package migration

import (
	"fmt"
	"os"

	"github.com/pelletier/go-toml/v2"
)

// WranglerConfig represents a subset of wrangler.toml
type WranglerConfig struct {
	Name               string            `toml:"name"`
	Main               string            `toml:"main"`
	CompatibilityDate  string            `toml:"compatibility_date"`
	CompatibilityFlags []string          `toml:"compatibility_flags"`
	D1Databases        []WranglerD1      `toml:"d1_databases"`
	KVNamespaces       []WranglerKV      `toml:"kv_namespaces"`
	Vars               map[string]string `toml:"vars"`
}

type WranglerD1 struct {
	Binding      string `toml:"binding"`
	DatabaseName string `toml:"database_name"`
	DatabaseID   string `toml:"database_id"`
}

type WranglerKV struct {
	Binding string `toml:"binding"`
	ID      string `toml:"id"`
}

// Detect looks for wrangler.toml in the current directory
func Detect(cwd string) (*WranglerConfig, error) {
	tomlPath := fmt.Sprintf("%s/wrangler.toml", cwd)
	if _, err := os.Stat(tomlPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("no wrangler.toml found in %s", cwd)
	}

	content, err := os.ReadFile(tomlPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read wrangler.toml: %w", err)
	}

	var config WranglerConfig
	if err := toml.Unmarshal(content, &config); err != nil {
		return nil, fmt.Errorf("failed to parse wrangler.toml: %w", err)
	}

	return &config, nil
}
