package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tmc/langchaingo/llms"
)

// Define tools as langchaingo.Tool
// We'll use function calling definition style

func (a *Agent) GetTools() []llms.Tool {
	return []llms.Tool{
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "read_file",
				Description: "Read the content of a file",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"path": map[string]interface{}{
							"type":        "string",
							"description": "The relative path to the file",
						},
					},
					"required": []string{"path"},
				},
			},
		},
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "list_dir",
				Description: "List files and directories in a path",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"path": map[string]interface{}{
							"type":        "string",
							"description": "The directory path to list (default: .)",
						},
					},
					"required": []string{"path"},
				},
			},
		},
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "search_symbols",
				Description: "Search for symbols (functions, classes) in the project",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"query": map[string]interface{}{
							"type":        "string",
							"description": "The name or kind of symbol to search for",
						},
					},
					"required": []string{"query"},
				},
			},
		},
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "write_file",
				Description: "Create or overwrite a file with new content",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"path": map[string]interface{}{
							"type":        "string",
							"description": "The relative path to the file",
						},
						"content": map[string]interface{}{
							"type":        "string",
							"description": "The content to write to the file",
						},
					},
					"required": []string{"path", "content"},
				},
			},
		},
	}
}

// validatePath ensures the requested path stays within the project root directory.
// It prevents path traversal attacks (e.g., "../../etc/passwd").
func validatePath(requestedPath string) (string, error) {
	projectRoot, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to determine project root: %w", err)
	}

	// Resolve to absolute path
	absPath := requestedPath
	if !filepath.IsAbs(absPath) {
		absPath = filepath.Join(projectRoot, absPath)
	}
	absPath = filepath.Clean(absPath)

	// Ensure the resolved path is within the project root
	if !strings.HasPrefix(absPath, projectRoot+string(filepath.Separator)) && absPath != projectRoot {
		return "", fmt.Errorf("access denied: path %q is outside project directory", requestedPath)
	}

	return absPath, nil
}

// ExecuteToolCall executes a requested tool call
func (a *Agent) ExecuteToolCall(ctx context.Context, name string, args string) (string, error) {
	var params map[string]interface{}
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	switch name {
	case "read_file":
		path, _ := params["path"].(string)
		return a.readFile(path)
	case "list_dir":
		path, _ := params["path"].(string)
		if path == "" {
			path = "."
		}
		return a.listDir(path)
	case "search_symbols":
		query, _ := params["query"].(string)
		return a.searchSymbols(query)
	case "write_file":
		path, _ := params["path"].(string)
		content, _ := params["content"].(string)
		return a.writeFile(path, content)
	default:
		return "", fmt.Errorf("unknown tool: %s", name)
	}
}

func (a *Agent) writeFile(path, content string) (string, error) {
	safePath, err := validatePath(path)
	if err != nil {
		return "", err
	}

	// Ensure directory exists
	dir := filepath.Dir(safePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(safePath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}
	return fmt.Sprintf("Successfully wrote %d bytes to %s", len(content), path), nil
}

func (a *Agent) readFile(path string) (string, error) {
	safePath, err := validatePath(path)
	if err != nil {
		return "", err
	}

	content, err := os.ReadFile(safePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	return string(content), nil
}

func (a *Agent) listDir(path string) (string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return "", fmt.Errorf("failed to list directory: %w", err)
	}

	var result string
	for _, entry := range entries {
		prefix := "F"
		if entry.IsDir() {
			prefix = "D"
		}
		info, err := entry.Info()
		if err != nil {
			result += fmt.Sprintf("[%s] %s (unknown size)\n", prefix, entry.Name())
			continue
		}
		result += fmt.Sprintf("[%s] %s (%d bytes)\n", prefix, entry.Name(), info.Size())
	}
	return result, nil
}

func (a *Agent) searchSymbols(query string) (string, error) {
	symbols, err := a.pkg.SearchSymbols(query)
	if err != nil {
		return "", fmt.Errorf("search failed: %w", err)
	}

	if len(symbols) == 0 {
		return "No symbols found.", nil
	}

	var result string
	for _, s := range symbols {
		result += fmt.Sprintf("%s (%s) in %s:%d\n", s.Name, s.Kind, s.FilePath, s.LineStart)
	}
	return result, nil
}
