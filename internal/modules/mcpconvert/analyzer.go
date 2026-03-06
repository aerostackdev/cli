package mcpconvert

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// PackageSource describes where the NPM package came from.
type PackageSource int

const (
	SourceNPM    PackageSource = iota // --package @org/mcp-server
	SourceGitHub                      // --github https://github.com/org/repo
	SourceLocal                       // --dir ./path/to/project
)

// AnalysisResult holds everything we learned about an MCP package.
type AnalysisResult struct {
	PackageName string   // e.g. "@notionhq/notion-mcp-server"
	EntryPoint  string   // resolved entry file path (relative)
	HasStdio    bool     // found StdioServerTransport usage
	HasHTTP     bool     // already uses HTTP transport
	EnvVars     []string // detected process.env.VAR_NAME references
	Warnings    []IncompatibleAPI
	ToolCount   int      // number of server.tool(...) calls found
	ProjectDir  string   // root directory of the extracted/cloned project
}

// IncompatibleAPI flags a Node.js API that doesn't work in Workers.
type IncompatibleAPI struct {
	API      string // e.g. "fs.readFile"
	Severity string // "warn" or "error"
	Message  string
}

// incompatRules maps Node.js module imports to severity + message.
var incompatRules = map[string]IncompatibleAPI{
	"child_process": {API: "child_process", Severity: "error", Message: "Cannot spawn subprocesses in Workers. Manual port required."},
	"fs":            {API: "fs", Severity: "warn", Message: "File system access not available in Workers. Tools using fs may fail at runtime."},
	"sqlite3":       {API: "sqlite3", Severity: "warn", Message: "Use Cloudflare D1 instead. Auto-conversion not possible for DB tools."},
	"better-sqlite3": {API: "better-sqlite3", Severity: "warn", Message: "Use Cloudflare D1 instead. Auto-conversion not possible for DB tools."},
	"net":           {API: "net", Severity: "warn", Message: "Raw TCP sockets not available in Workers."},
	"dgram":         {API: "dgram", Severity: "warn", Message: "UDP sockets not available in Workers."},
	"cluster":       {API: "cluster", Severity: "error", Message: "Cluster module not available in Workers."},
	"worker_threads": {API: "worker_threads", Severity: "warn", Message: "Worker threads not available. Workers are already single-threaded."},
}

// Analyze inspects the given project directory for MCP patterns.
func Analyze(projectDir string) (*AnalysisResult, error) {
	result := &AnalysisResult{
		ProjectDir: projectDir,
	}

	// 1. Read package.json to find entry point + name
	pkgPath := filepath.Join(projectDir, "package.json")
	pkgData, err := os.ReadFile(pkgPath)
	if err != nil {
		return nil, fmt.Errorf("no package.json found: %w", err)
	}

	var pkg struct {
		Name string            `json:"name"`
		Main string            `json:"main"`
		Bin  json.RawMessage   `json:"bin"`
	}
	if err := json.Unmarshal(pkgData, &pkg); err != nil {
		return nil, fmt.Errorf("invalid package.json: %w", err)
	}
	result.PackageName = pkg.Name

	// 2. Resolve entry point: bin > main > src/index.ts > index.ts
	entry := resolveEntry(pkg.Bin, pkg.Main, projectDir)
	if entry == "" {
		return nil, fmt.Errorf("cannot find entry point. Set 'main' or 'bin' in package.json")
	}
	result.EntryPoint = entry

	// 3. Scan all .ts/.js files for patterns
	err = filepath.Walk(projectDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		// Skip node_modules, dist, .git
		if info.IsDir() {
			base := info.Name()
			if base == "node_modules" || base == "dist" || base == ".git" || base == ".aerostack" {
				return filepath.SkipDir
			}
			return nil
		}
		ext := filepath.Ext(path)
		if ext != ".ts" && ext != ".js" && ext != ".mts" && ext != ".mjs" {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		src := string(content)

		// Check for transport types
		if strings.Contains(src, "StdioServerTransport") {
			result.HasStdio = true
		}
		if strings.Contains(src, "StreamableHTTPServerTransport") || strings.Contains(src, "SSEServerTransport") {
			result.HasHTTP = true
		}

		// Count tool definitions
		toolRe := regexp.MustCompile(`\.tool\s*\(`)
		result.ToolCount += len(toolRe.FindAllString(src, -1))

		// Detect process.env references
		envRe := regexp.MustCompile(`process\.env\.([A-Z][A-Z0-9_]*)`)
		for _, match := range envRe.FindAllStringSubmatch(src, -1) {
			if len(match) > 1 {
				result.EnvVars = appendUnique(result.EnvVars, match[1])
			}
		}

		// Detect process.env["KEY"] and process.env['KEY']
		envBracketRe := regexp.MustCompile(`process\.env\[['"]([A-Z][A-Z0-9_]*)['"]\]`)
		for _, match := range envBracketRe.FindAllStringSubmatch(src, -1) {
			if len(match) > 1 {
				result.EnvVars = appendUnique(result.EnvVars, match[1])
			}
		}

		// Check for incompatible APIs
		for mod, rule := range incompatRules {
			// Check import/require
			importRe := regexp.MustCompile(fmt.Sprintf(`(?:from\s+['"]%s['"]|require\s*\(\s*['"]%s['"]\s*\))`, regexp.QuoteMeta(mod), regexp.QuoteMeta(mod)))
			if importRe.MatchString(src) {
				result.Warnings = append(result.Warnings, rule)
			}
		}

		return nil
	})

	return result, err
}

// resolveEntry finds the best entry point from package.json fields.
func resolveEntry(binRaw json.RawMessage, main string, dir string) string {
	// Try bin field first (can be string or object)
	if len(binRaw) > 0 {
		var binStr string
		if json.Unmarshal(binRaw, &binStr) == nil && binStr != "" {
			return binStr
		}
		var binMap map[string]string
		if json.Unmarshal(binRaw, &binMap) == nil {
			// Take first value
			for _, v := range binMap {
				return v
			}
		}
	}

	// Try main field
	if main != "" {
		return main
	}

	// Fallback: check common entry points
	candidates := []string{
		"src/index.ts",
		"src/index.mts",
		"index.ts",
		"src/index.js",
		"index.js",
	}
	for _, c := range candidates {
		if _, err := os.Stat(filepath.Join(dir, c)); err == nil {
			return c
		}
	}

	return ""
}

func appendUnique(slice []string, val string) []string {
	for _, s := range slice {
		if s == val {
			return slice
		}
	}
	return append(slice, val)
}
