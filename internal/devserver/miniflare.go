package devserver

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// Service represents a multi-worker service from [[services]]
type Service struct {
	Name string
	Main string
}

// AerostackConfig represents key fields from aerostack.toml
type AerostackConfig struct {
	Name              string
	Main              string
	CompatibilityDate string
	D1Databases       []D1Database
	PostgresDatabases []PostgresDatabase
	// EnvOverrides: env-specific D1 database_id overrides (from [env.staging], [env.production])
	EnvOverrides map[string][]D1Database
	// Services: multi-worker services from [[services]]
	Services []Service
}

// D1Database represents a D1 database binding
type D1Database struct {
	Binding      string
	DatabaseName string
	DatabaseID   string
}

// PostgresDatabase represents an external Postgres database binding
type PostgresDatabase struct {
	Binding          string
	ConnectionString string
	Schema           string // Optional schema file path
	PoolSize         int    // Connection pool size
}

// ParseAerostackToml reads aerostack.toml and extracts config (simple parsing to avoid heavy deps)
func ParseAerostackToml(path string) (*AerostackConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read aerostack.toml: %w", err)
	}
	content := string(data)

	cfg := &AerostackConfig{
		CompatibilityDate: "2024-01-01",
		D1Databases:       []D1Database{},
		PostgresDatabases: []PostgresDatabase{},
		EnvOverrides:      map[string][]D1Database{},
	}

	// Parse env-specific overrides ([env.staging], [env.production])
	cfg.EnvOverrides["staging"] = parseEnvD1Databases(content, "env.staging")
	cfg.EnvOverrides["production"] = parseEnvD1Databases(content, "env.production")

	// Simple key = "value" extraction
	cfg.Name = extractTomlString(content, "name")
	cfg.Main = extractTomlString(content, "main")
	if d := extractTomlString(content, "compatibility_date"); d != "" {
		cfg.CompatibilityDate = d
	}

	// Parse [[d1_databases]] blocks
	cfg.D1Databases = parseD1Databases(content)

	// Parse [[postgres_databases]] blocks
	cfg.PostgresDatabases = parsePostgresDatabases(content)

	// Parse [[services]] blocks (multi-worker)
	cfg.Services = parseServices(content)

	// Defaults
	if cfg.Main == "" {
		cfg.Main = "src/index.ts"
	}
	if cfg.Name == "" {
		cfg.Name = "aerostack-app"
	}

	return cfg, nil
}

func extractTomlString(content, key string) string {
	re := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(key) + `\s*=\s*"([^"]*)"`)
	m := re.FindStringSubmatch(content)
	if len(m) > 1 {
		return m[1]
	}
	return ""
}

// parseEnvD1Databases parses [[env.staging.d1_databases]] or [[env.production.d1_databases]] blocks
func parseEnvD1Databases(content string, envPrefix string) []D1Database {
	var dbs []D1Database
	blockRe := regexp.MustCompile(`\[\[` + regexp.QuoteMeta(envPrefix) + `\.d1_databases\]\]\s*\n([\s\S]*?)(?:\n\[\[|\n\[|\z)`)
	blocks := blockRe.FindAllStringSubmatch(content, -1)
	for _, block := range blocks {
		if len(block) < 2 {
			continue
		}
		inner := block[1]
		binding := extractTomlString(inner, "binding")
		dbName := extractTomlString(inner, "database_name")
		dbID := extractTomlString(inner, "database_id")
		if binding != "" && dbID != "" {
			if dbName == "" {
				dbName = "local-db"
			}
			dbs = append(dbs, D1Database{Binding: binding, DatabaseName: dbName, DatabaseID: dbID})
		}
	}
	return dbs
}

func parseD1Databases(content string) []D1Database {
	var dbs []D1Database
	// Match [[d1_databases]] ... binding = "X" ... database_name = "Y" ... database_id = "Z"
	blockRe := regexp.MustCompile(`\[\[d1_databases\]\]\s*\n([\s\S]*?)(?:\n\[\[|\n\[|\z)`)
	blocks := blockRe.FindAllStringSubmatch(content, -1)
	for _, block := range blocks {
		if len(block) < 2 {
			continue
		}
		inner := block[1]
		binding := extractTomlString(inner, "binding")
		dbName := extractTomlString(inner, "database_name")
		dbID := extractTomlString(inner, "database_id")
		if binding != "" {
			if dbID == "" {
				dbID = "aerostack-local"
			}
			if dbName == "" {
				dbName = "local-db"
			}
			dbs = append(dbs, D1Database{Binding: binding, DatabaseName: dbName, DatabaseID: dbID})
		}
	}
	return dbs
}

func parsePostgresDatabases(content string) []PostgresDatabase {
	var dbs []PostgresDatabase
	// Match [[postgres_databases]] blocks
	blockRe := regexp.MustCompile(`\[\[postgres_databases\]\]\s*\n([\s\S]*?)(?:\n\[\[|\n\[|\z)`)
	blocks := blockRe.FindAllStringSubmatch(content, -1)
	for _, block := range blocks {
		if len(block) < 2 {
			continue
		}
		inner := block[1]
		binding := extractTomlString(inner, "binding")
		connStr := extractTomlString(inner, "connection_string")
		schema := extractTomlString(inner, "schema")
		poolSize := extractTomlInt(inner, "pool_size")

		if binding != "" && connStr != "" {
			if poolSize == 0 {
				poolSize = 10 // Default pool size
			}

			// Interpolate environment variables ($VAR_NAME or ${VAR_NAME})
			connStr = interpolateEnvVars(connStr)

			dbs = append(dbs, PostgresDatabase{
				Binding:          binding,
				ConnectionString: connStr,
				Schema:           schema,
				PoolSize:         poolSize,
			})
		}
	}
	return dbs
}

// interpolateEnvVars replaces $VAR_NAME or ${VAR_NAME} with environment variable values
func interpolateEnvVars(s string) string {
	// Replace ${VAR_NAME}
	s = regexp.MustCompile(`\$\{([A-Z_][A-Z0-9_]*)\}`).ReplaceAllStringFunc(s, func(match string) string {
		varName := match[2 : len(match)-1] // Strip ${ and }
		if val := os.Getenv(varName); val != "" {
			return val
		}
		return match // Keep original if not found
	})

	// Replace $VAR_NAME
	s = regexp.MustCompile(`\$([A-Z_][A-Z0-9_]*)`).ReplaceAllStringFunc(s, func(match string) string {
		varName := match[1:] // Strip $
		if val := os.Getenv(varName); val != "" {
			return val
		}
		return match // Keep original if not found
	})

	return s
}

// parseServices parses [[services]] blocks for multi-worker dev
func parseServices(content string) []Service {
	var svcs []Service
	blockRe := regexp.MustCompile(`\[\[services\]\]\s*\n([\s\S]*?)(?:\n\[\[|\n\[|\z)`)
	blocks := blockRe.FindAllStringSubmatch(content, -1)
	for _, block := range blocks {
		if len(block) < 2 {
			continue
		}
		inner := block[1]
		name := extractTomlString(inner, "name")
		main := extractTomlString(inner, "main")
		if name != "" && main != "" {
			svcs = append(svcs, Service{Name: name, Main: main})
		}
	}
	return svcs
}

func extractTomlInt(content, key string) int {
	re := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(key) + `\s*=\s*(\d+)`)
	m := re.FindStringSubmatch(content)
	if len(m) > 1 {
		val, _ := strconv.Atoi(m[1])
		return val
	}
	return 0
}

// GenerateWranglerToml creates wrangler.toml from AerostackConfig
func GenerateWranglerToml(cfg *AerostackConfig, outputPath string) error {
	var sb strings.Builder
	// Use dist/worker.js when we have a build step (@shared alias)
	buildCmd := fmt.Sprintf("npx esbuild %q --bundle --outfile=dist/worker.js --format=esm --alias:@shared=./shared", cfg.Main)
	sb.WriteString(fmt.Sprintf("name = %q\n", cfg.Name))
	sb.WriteString("main = \"dist/worker.js\"\n")
	sb.WriteString(fmt.Sprintf("compatibility_date = %q\n\n", cfg.CompatibilityDate))
	sb.WriteString("# @shared alias: import from shared/ via 'import x from \"@shared/db\"'\n")
	sb.WriteString("[build]\n")
	sb.WriteString(fmt.Sprintf("command = %q\n\n", buildCmd))

	for _, db := range cfg.D1Databases {
		sb.WriteString("[[d1_databases]]\n")
		sb.WriteString(fmt.Sprintf("binding = %q\n", db.Binding))
		sb.WriteString(fmt.Sprintf("database_name = %q\n", db.DatabaseName))
		sb.WriteString(fmt.Sprintf("database_id = %q\n\n", db.DatabaseID))
	}

	// Hyperdrive bindings for Postgres (local: set CLOUDFLARE_HYPERDRIVE_LOCAL_CONNECTION_STRING_<BINDING>; remote: add id from wrangler hyperdrive create)
	for _, pg := range cfg.PostgresDatabases {
		sb.WriteString("[[hyperdrive]]\n")
		sb.WriteString(fmt.Sprintf("binding = %q\n", pg.Binding))
		sb.WriteString("# For local: set CLOUDFLARE_HYPERDRIVE_LOCAL_CONNECTION_STRING_" + pg.Binding + " in .env\n")
		sb.WriteString("# For remote: run 'wrangler hyperdrive create <name> --connection-string=...' and add id here\n\n")
	}

	// Env blocks for deploy --env staging/production (use overrides from aerostack.toml if present)
	sb.WriteString("# Deploy: aerostack deploy --env staging | production\n")
	for _, envName := range []string{"staging", "production"} {
		sb.WriteString(fmt.Sprintf("[env.%s]\n", envName))
		dbs := cfg.EnvOverrides[envName]
		if len(dbs) == 0 {
			dbs = cfg.D1Databases
		}
		for _, db := range dbs {
			sb.WriteString(fmt.Sprintf("[[env.%s.d1_databases]]\nbinding = %q\ndatabase_name = %q\ndatabase_id = %q\n\n", envName, db.Binding, db.DatabaseName, db.DatabaseID))
		}
	}

	if err := os.WriteFile(outputPath, []byte(sb.String()), 0644); err != nil {
		return fmt.Errorf("failed to write wrangler.toml: %w", err)
	}
	return nil
}

// GenerateWranglerTomlForService creates a wrangler.toml for a specific service (multi-worker)
func GenerateWranglerTomlForService(cfg *AerostackConfig, svc Service, outputPath string) error {
	outfile := fmt.Sprintf("dist/%s.js", svc.Name)
	buildCmd := fmt.Sprintf("npx esbuild %q --bundle --outfile=%s --format=esm --alias:@shared=./shared", svc.Main, outfile)
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("name = %q\n", cfg.Name+"-"+svc.Name))
	sb.WriteString(fmt.Sprintf("main = %q\n", outfile))
	sb.WriteString(fmt.Sprintf("compatibility_date = %q\n\n", cfg.CompatibilityDate))
	sb.WriteString("[build]\n")
	sb.WriteString(fmt.Sprintf("command = %q\n\n", buildCmd))
	for _, db := range cfg.D1Databases {
		sb.WriteString("[[d1_databases]]\n")
		sb.WriteString(fmt.Sprintf("binding = %q\n", db.Binding))
		sb.WriteString(fmt.Sprintf("database_name = %q\n", db.DatabaseName))
		sb.WriteString(fmt.Sprintf("database_id = %q\n\n", db.DatabaseID))
	}
	for _, pg := range cfg.PostgresDatabases {
		sb.WriteString("[[hyperdrive]]\n")
		sb.WriteString(fmt.Sprintf("binding = %q\n", pg.Binding))
		sb.WriteString("# Set CLOUDFLARE_HYPERDRIVE_LOCAL_CONNECTION_STRING_" + pg.Binding + " in .env\n\n")
	}
	if err := os.WriteFile(outputPath, []byte(sb.String()), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", outputPath, err)
	}
	return nil
}

// CheckNode checks if Node.js 18+ is installed
func CheckNode() (version string, err error) {
	cmd := exec.Command("node", "-v")
	cmd.Env = os.Environ()
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("Node.js is required for 'aerostack dev' (D1 support). Install Node.js 18+ from https://nodejs.org")
	}
	version = strings.TrimSpace(string(out))
	// Parse v18.0.0 -> 18
	verStr := strings.TrimPrefix(version, "v")
	parts := strings.Split(verStr, ".")
	if len(parts) < 1 {
		return version, nil
	}
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return version, nil
	}
	if major < 18 {
		return "", fmt.Errorf("Node.js 18+ required (found %s). Install from https://nodejs.org", version)
	}
	return version, nil
}

// RunWranglerDev runs npx wrangler dev with the given config
// remoteEnv: if non-empty, passes --remote to use real Cloudflare bindings (e.g. "staging")
// hyperdriveEnvVars: optional map of env var name -> value for Hyperdrive local connection strings
func RunWranglerDev(wranglerTomlPath string, port int, remoteEnv string, hyperdriveEnvVars map[string]string) (*exec.Cmd, error) {
	absPath, err := filepath.Abs(wranglerTomlPath)
	if err != nil {
		return nil, err
	}
	projectRoot := filepath.Dir(absPath)

	args := []string{"-y", "wrangler@latest", "dev", "--config", absPath, "--port", strconv.Itoa(port)}
	if remoteEnv != "" {
		args = append(args, "--remote", "--env", remoteEnv)
	}
	cmd := exec.Command("npx", args...)
	cmd.Dir = projectRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	// Ensure npx finds packages (use project dir)
	cmd.Env = append(cmd.Env, "NPX_UPDATE_NOTIFIER=false")

	// Inject Hyperdrive local connection strings (avoids writing secrets to wrangler.toml)
	for k, v := range hyperdriveEnvVars {
		cmd.Env = append(cmd.Env, k+"="+v)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start wrangler: %w", err)
	}
	return cmd, nil
}

// EnsureDefaultD1 adds a default D1 binding if none exist (for blank template)
func EnsureDefaultD1(cfg *AerostackConfig) {
	if len(cfg.D1Databases) == 0 {
		cfg.D1Databases = []D1Database{
			{Binding: "DB", DatabaseName: "local-db", DatabaseID: "aerostack-local"},
		}
	}
}
