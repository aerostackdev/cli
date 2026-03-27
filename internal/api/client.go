package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

var httpClient = &http.Client{
	Timeout: 30 * time.Second,
}

// httpClientLong is used for operations that deploy Workers (CF API can take 60-90s)
var httpClientLong = &http.Client{
	Timeout: 120 * time.Second,
}

func getBaseURL() string {
	return BaseURL()
}

// BaseURL returns the Aerostack API base URL (env AEROSTACK_API_URL or default).
func BaseURL() string {
	if u := os.Getenv("AEROSTACK_API_URL"); u != "" {
		return u
	}
	return "https://api.aerostack.dev"
}

type ValidateResponse struct {
	KeyType     string `json:"keyType"`     // "account" or "project"
	ProjectID   string `json:"projectId"`   // project key only
	ProjectName string `json:"projectName"` // project key only
	Slug        string `json:"slug"`        // project key only
	URL         string `json:"url"`         // project key only
	UserID      string `json:"userId"`      // account key only
	Email       string `json:"email"`       // account key only
	Name        string `json:"name"`        // account key only
	Message     string `json:"message"`     // account key hint
}

type ProjectMetadata struct {
	ProjectID   string `json:"projectId"`
	Name        string `json:"name"`
	Collections []struct {
		ID                string  `json:"id"`
		Name              string  `json:"name"`
		Slug              string  `json:"slug"`
		SchemaComponentID string  `json:"schema_component_id"`
		Schema            *string `json:"schema"` // JSON string
	} `json:"collections"`
	Hooks []struct {
		ID             string   `json:"id"`
		Name           string   `json:"name"`
		Slug           string   `json:"slug"`
		EventType      string   `json:"event_type"`
		Type           string   `json:"type"`
		IsPublic       int      `json:"is_public"`
		AllowedMethods []string `json:"allowed_methods"`
	} `json:"hooks"`
}

type ValidateError struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func Validate(apiKey string) (*ValidateResponse, error) {
	url := getBaseURL() + "/api/v1/cli/validate"

	// Debug logging for 401 troubleshooting
	if os.Getenv("DEBUG") == "true" {
		prefix := "none"
		if len(apiKey) > 4 {
			prefix = apiKey[:4] + "..."
		}
		fmt.Printf("[DEBUG] Validating key with prefix %s at %s\n", prefix, url)
	}

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Key", apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		var errBody ValidateError
		_ = json.Unmarshal(body, &errBody)
		msg := errBody.Error.Message
		if msg == "" {
			msg = string(body)
		}
		return nil, fmt.Errorf("validate failed (%d): %s", resp.StatusCode, msg)
	}

	var out ValidateResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type TelemetryPayload struct {
	ProjectID    string   `json:"projectId"`
	Command      string   `json:"command"`
	ErrorMessage string   `json:"error_message"`
	ErrorStack   string   `json:"error_stack"`
	Logs         []string `json:"logs"`
	OS           string   `json:"os"`
	CLIVersion   string   `json:"cli_version"`
}

func SendTelemetry(apiKey string, payload TelemetryPayload) error {
	if apiKey == "" {
		return fmt.Errorf("API key required for telemetry")
	}

	url := getBaseURL() + "/api/v1/cli/telemetry/errors"
	jsonBody, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return err
	}

	req.Header.Set("X-API-Key", apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return fmt.Errorf("telemetry failed with status %d", resp.StatusCode)
	}

	return nil
}

func GetProjectMetadata(apiKey string, projectSlug string) (*ProjectMetadata, error) {
	url := getBaseURL() + "/api/v1/cli/project-metadata"
	if projectSlug != "" {
		url += "?project=" + projectSlug
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Key", apiKey)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("metadata fetch failed (%d): %s", resp.StatusCode, string(body))
	}

	var out ProjectMetadata
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type DeployResponse struct {
	Success    bool   `json:"success"`
	ScriptName string `json:"scriptName"`
	URL        string `json:"url"`
	Project    struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Slug string `json:"slug"`
	} `json:"project"`
	PublicURL string `json:"publicUrl"`
	Env       string `json:"env"`
	IsPublic  bool   `json:"isPublic"`
}

type CommunityFunction struct {
	ID             string            `json:"id"`
	Slug           string            `json:"slug"`
	Name           string            `json:"name"`
	Description    string            `json:"description"`
	Readme         string            `json:"readme"`
	Category       string            `json:"category"`
	Tags           []string          `json:"tags"`
	Language       string            `json:"language"`
	Runtime        string            `json:"runtime"`
	Code           string            `json:"code"`
	Files          map[string]string `json:"files,omitempty"`
	ConfigSchema   interface{}       `json:"config_schema"`
	License        string            `json:"license"`
	Version        string            `json:"version"`
	Status         string            `json:"status"`
	AuthorUsername string            `json:"author_username"`
	URL            string            `json:"url"`
}

// InstallFile is a single source file in an OFS install manifest.
type InstallFile struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// InstallManifest is the full manifest returned by the registry install endpoint.
// It carries all files needed to wire a function into a consumer project.
type InstallManifest struct {
	Name            string        `json:"name"`
	Slug            string        `json:"slug"`
	Description     string        `json:"description"`
	Author          string        `json:"author"`
	Version         string        `json:"version"`
	Category        string        `json:"category"`
	Language        string        `json:"language"`
	License         string        `json:"license"`
	StarCount       int           `json:"starCount"`
	CloneCount      int           `json:"cloneCount"`
	Tags            []string      `json:"tags"`
	NpmDependencies []string      `json:"npmDependencies"`
	EnvVars         []string      `json:"envVars"`
	RouteExport     string        `json:"routeExport"`
	RoutePath       string        `json:"routePath"`
	DrizzleSchema   bool          `json:"drizzleSchema"`
	Files           []InstallFile `json:"files"`
}

type CommunityPushResponse struct {
	ID     string `json:"id"`
	Slug   string `json:"slug"`
	Author string `json:"author"`
	Status string `json:"status"`
	URL    string `json:"url"`
}

type DeployError struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Details string `json:"details"`
	} `json:"error"`
}

func Deploy(apiKey string, files map[string]string, env string, projectName string, projectID string, isPublic bool, isPrivate bool, bindingsJSON string, compatDate string, compatFlags []string) (*DeployResponse, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	_ = w.WriteField("env", env)
	if projectID != "" {
		_ = w.WriteField("project_id", projectID)
	}
	if projectName != "" {
		_ = w.WriteField("name", projectName)
	}
	if bindingsJSON != "" {
		_ = w.WriteField("bindings", bindingsJSON)
	}
	// Add is_public field if explicitly set
	if isPublic {
		_ = w.WriteField("isPublic", "true")
	} else if isPrivate {
		_ = w.WriteField("isPublic", "false")
	}

	if compatDate != "" {
		_ = w.WriteField("compatibility_date", compatDate)
	}
	if len(compatFlags) > 0 {
		flagsJSON, _ := json.Marshal(compatFlags)
		_ = w.WriteField("compatibility_flags", string(flagsJSON))
	}

	for name, path := range files {
		workerData, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read file %s: %w", path, err)
		}
		part, err := w.CreateFormFile(name, name)
		if err != nil {
			return nil, err
		}
		if _, err := part.Write(workerData); err != nil {
			return nil, err
		}
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	url := getBaseURL() + "/api/v1/cli/deploy"
	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Key", apiKey)
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		var errBody DeployError
		_ = json.Unmarshal(body, &errBody)
		msg := errBody.Error.Message
		if msg == "" {
			msg = string(body)
		}
		if errBody.Error.Details != "" {
			msg = fmt.Sprintf("%s - %s", msg, errBody.Error.Details)
		}
		return nil, fmt.Errorf("deploy failed (%d): %s", resp.StatusCode, msg)
	}

	var out DeployResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type CreateProjectResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

func CreateProject(apiKey, name string) (*CreateProjectResponse, error) {
	url := getBaseURL() + "/api/v1/cli/projects"

	bodyData := map[string]string{"name": name}
	jsonBody, _ := json.Marshal(bodyData)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Key", apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		var errBody struct {
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		_ = json.Unmarshal(body, &errBody)
		msg := errBody.Error.Message
		if msg == "" {
			msg = string(body)
		}
		return nil, fmt.Errorf("create project failed (%d): %s", resp.StatusCode, msg)
	}

	var out CreateProjectResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type ProjectListItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
	URL  string `json:"url"`
}

func ListProjects(apiKey string) ([]ProjectListItem, error) {
	url := getBaseURL() + "/api/v1/cli/projects"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Key", apiKey)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		var errBody struct {
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		_ = json.Unmarshal(body, &errBody)
		msg := errBody.Error.Message
		if msg == "" {
			msg = string(body)
		}
		return nil, fmt.Errorf("list projects failed (%d): %s", resp.StatusCode, msg)
	}

	var out struct {
		Projects []ProjectListItem `json:"projects"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	return out.Projects, nil
}

// ─── Community Functions ───────────────────────────────────────────────────

func CommunityPush(apiKey string, fn CommunityFunction) (*CommunityPushResponse, error) {
	url := getBaseURL() + "/api/community/functions"
	jsonBody, _ := json.Marshal(fn)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Key", apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return nil, fmt.Errorf("push failed (%d): %s", resp.StatusCode, string(body))
	}

	var out CommunityPushResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func CommunityPull(username, slug string) (*CommunityFunction, error) {
	url := fmt.Sprintf("%s/api/community/functions/%s/%s", getBaseURL(), username, slug)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("pull failed (%d): %s", resp.StatusCode, string(body))
	}

	var out struct {
		Function CommunityFunction `json:"function"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	return &out.Function, nil
}

func CommunityPullNoAuth(slug string) (*CommunityFunction, error) {
	url := fmt.Sprintf("%s/api/community/functions/install/%s", getBaseURL(), slug)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("install search failed (%d): %s", resp.StatusCode, string(body))
	}

	var out struct {
		Function CommunityFunction `json:"function"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	return &out.Function, nil
}

func CommunityPublish(apiKey, id string) error {
	url := fmt.Sprintf("%s/api/community/functions/%s/publish", getBaseURL(), id)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-API-Key", apiKey)

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("publish failed (%d): %s", resp.StatusCode, string(body))
	}
	return nil
}

type CommunityDeployMcpResponse struct {
	Success     bool   `json:"success"`
	Hosted      bool   `json:"hosted"`
	WorkerURL   string `json:"url"`
	MCPServerID string `json:"mcp_server_id"`
	Slug        string `json:"slug"`
	Env         string `json:"env"`
	ToolsCount  int    `json:"tools_count"`
	Message     string `json:"message"`
}

func CommunityDeployMcp(apiKey string, workerPath string, slug string, env string, envVars []string, description string, category string, tags string) (*CommunityDeployMcpResponse, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	_ = w.WriteField("slug", slug)
	_ = w.WriteField("env", env)
	if len(envVars) > 0 {
		envVarsJSON, _ := json.Marshal(envVars)
		_ = w.WriteField("env_vars", string(envVarsJSON))
	}
	if description != "" {
		_ = w.WriteField("description", description)
	}
	if category != "" {
		_ = w.WriteField("category", category)
	}
	if tags != "" {
		_ = w.WriteField("tags", tags)
	}

	workerData, err := os.ReadFile(workerPath)
	if err != nil {
		return nil, fmt.Errorf("read worker file: %w", err)
	}
	part, err := w.CreateFormFile("worker", "worker.js")
	if err != nil {
		return nil, err
	}
	if _, err := part.Write(workerData); err != nil {
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	url := getBaseURL() + "/api/v1/cli/deploy/mcp"
	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Key", apiKey)
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := httpClientLong.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 && resp.StatusCode != 201 && resp.StatusCode != 210 {
		return nil, fmt.Errorf("mcp deploy failed (%d): %s", resp.StatusCode, string(body))
	}

	var out CommunityDeployMcpResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type CommunityDeploySkillResponse struct {
	Success     bool   `json:"success"`
	URL         string `json:"url"`
	Slug        string `json:"slug"`
	Env         string `json:"env"`
	Message     string `json:"message"`
	FunctionURL string `json:"function_url,omitempty"`
	FunctionID  string `json:"function_id,omitempty"`
}

func CommunityDeploySkill(apiKey string, name string, content string, functionCode string, env string) (*CommunityDeploySkillResponse, error) {
	url := getBaseURL() + "/api/v1/cli/deploy/skill"

	payload := map[string]string{
		"name":          name,
		"content":       content,
		"function_code": functionCode,
		"env":           env,
	}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Key", apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return nil, fmt.Errorf("skill deploy failed (%d): %s", resp.StatusCode, string(body))
	}

	var out CommunityDeploySkillResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CommunityGetInstallManifest fetches the full OFS install manifest for username/slug.
// It calls GET /api/community/functions/:username/:slug/install which returns the
// multi-file manifest with routeExport, routePath, npmDependencies, etc.
func CommunityGetInstallManifest(username, slug string) (*InstallManifest, error) {
	url := fmt.Sprintf("%s/api/community/functions/%s/%s/install", getBaseURL(), username, slug)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("install manifest fetch failed (%d): %s", resp.StatusCode, string(body))
	}

	var manifest InstallManifest
	if err := json.Unmarshal(body, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse install manifest: %w", err)
	}
	return &manifest, nil
}

// CommunityGetInstallManifestBySlug fetches the full OFS install manifest by slug only
// (picks the best match). Calls GET /api/community/functions/install/:slug.
func CommunityGetInstallManifestBySlug(slug string) (*InstallManifest, error) {
	url := fmt.Sprintf("%s/api/community/functions/install/%s", getBaseURL(), slug)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("install manifest fetch failed (%d): %s", resp.StatusCode, string(body))
	}

	var manifest InstallManifest
	if err := json.Unmarshal(body, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse install manifest: %w", err)
	}
	return &manifest, nil
}

// ─── Skill types ──────────────────────────────────────────────────────────────

type SkillInfo struct {
	ID             string `json:"id"`
	Slug           string `json:"slug"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	Visibility     string `json:"visibility"`
	AuthorUsername string `json:"author_username"`
	AccessType     string `json:"access_type"`
	Tools          []struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"tools"`
}

type skillGetResponse struct {
	Server SkillInfo `json:"server"`
}

type SkillPublishPayload struct {
	Name               string   `json:"name"`
	Description        string   `json:"description"`
	Tools              []any    `json:"tools,omitempty"`
	BackedByFunctionID string   `json:"backed_by_function_id,omitempty"`
	WorkerURL          string   `json:"worker_url,omitempty"`
	Tags               []string `json:"tags,omitempty"`
	Visibility         string   `json:"visibility,omitempty"`
	Publish            bool     `json:"publish,omitempty"`
}

type SkillPublishResponse struct {
	ID     string `json:"id"`
	Slug   string `json:"slug"`
	Author string `json:"author"`
	Status string `json:"status"`
	URL    string `json:"url"`
}

// ─── Workspace types ──────────────────────────────────────────────────────────

type Workspace struct {
	ID         string `json:"id"`
	Slug       string `json:"slug"`
	Name       string `json:"name"`
	GatewayURL string `json:"gateway_url"`
}

type WorkspaceServer struct {
	WsServerID string `json:"ws_server_id"`
	ServerID   string `json:"server_id"`
	Slug       string `json:"slug"`
	Name       string `json:"name"`
	ToolCount  int    `json:"tool_count"`
	Enabled    bool   `json:"enabled"`
}

type WorkspaceDetail struct {
	Workspace
	Servers []WorkspaceServer `json:"servers"`
}

type WorkspaceListResponse struct {
	Workspaces []Workspace `json:"workspaces"`
}

// ─── Skill API functions ──────────────────────────────────────────────────────

// SkillGet fetches metadata for a skill by username/slug (public, no auth needed).
func SkillGet(username, slug string) (*SkillInfo, error) {
	url := fmt.Sprintf("%s/api/community/mcp/%s/%s", getBaseURL(), username, slug)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("skill '%s/%s' not found", username, slug)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to fetch skill (%d): %s", resp.StatusCode, string(body))
	}

	var wrapper skillGetResponse
	if err := json.Unmarshal(body, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse skill info: %w", err)
	}
	return &wrapper.Server, nil
}

// SkillPublish creates or updates a skill in the registry.
func SkillPublish(apiKey string, payload SkillPublishPayload) (*SkillPublishResponse, error) {
	url := getBaseURL() + "/api/community/mcp"
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", apiKey)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return nil, fmt.Errorf("skill publish failed (%d): %s", resp.StatusCode, string(body))
	}

	var result SkillPublishResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse publish response: %w", err)
	}
	return &result, nil
}

// ─── Workspace API functions ──────────────────────────────────────────────────

// WorkspaceList returns all workspaces owned by the authenticated user.
func WorkspaceList(apiKey string) ([]Workspace, error) {
	url := getBaseURL() + "/api/community/mcp/workspaces"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Key", apiKey)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("workspace list failed (%d): %s", resp.StatusCode, string(body))
	}

	var result WorkspaceListResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return result.Workspaces, nil
}

// WorkspaceCreate creates a new workspace and returns it.
func WorkspaceCreate(apiKey, name string) (*Workspace, error) {
	url := getBaseURL() + "/api/community/mcp/workspaces"
	data, _ := json.Marshal(map[string]string{"name": name})

	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", apiKey)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return nil, fmt.Errorf("workspace create failed (%d): %s", resp.StatusCode, string(body))
	}

	var ws Workspace
	if err := json.Unmarshal(body, &ws); err != nil {
		return nil, err
	}
	return &ws, nil
}

// WorkspaceGet returns full workspace details including servers.
func WorkspaceGet(apiKey, workspaceID string) (*WorkspaceDetail, error) {
	url := fmt.Sprintf("%s/api/community/mcp/workspaces/%s", getBaseURL(), workspaceID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Key", apiKey)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("workspace get failed (%d): %s", resp.StatusCode, string(body))
	}

	var detail WorkspaceDetail
	if err := json.Unmarshal(body, &detail); err != nil {
		return nil, err
	}
	return &detail, nil
}

// WorkspaceAddServer adds a skill/MCP server to a workspace by workspace ID + server ID.
func WorkspaceAddServer(apiKey, workspaceID, serverID string) error {
	url := fmt.Sprintf("%s/api/community/mcp/workspaces/%s/servers", getBaseURL(), workspaceID)
	data, _ := json.Marshal(map[string]string{"server_id": serverID})

	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", apiKey)

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == 409 {
		return fmt.Errorf("skill is already in this workspace")
	}
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return fmt.Errorf("failed to add skill to workspace (%d): %s", resp.StatusCode, string(body))
	}
	return nil
}

// WorkspaceTestTools calls tools/list on a workspace and returns the tools.
func WorkspaceTestTools(apiKey, workspaceID string) ([]WorkspaceToolInfo, error) {
	url := fmt.Sprintf("%s/api/community/mcp/workspaces/%s/tools", getBaseURL(), workspaceID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Key", apiKey)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("workspace test failed (%d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Tools []WorkspaceToolInfo `json:"tools"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return result.Tools, nil
}

// WorkspaceToolInfo holds info about a tool in a workspace.
type WorkspaceToolInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ServerSlug  string `json:"server_slug"`
}

// ─── MCP Pull ─────────────────────────────────────────────────────────────────

// McpPullResponse holds the files returned by GET /api/v1/cli/mcp/:slug.
type McpPullResponse struct {
	Slug          string `json:"slug"`
	Name          string `json:"name"`
	SrcIndexTs    string `json:"src_index_ts"`
	AerostackToml string `json:"aerostack_toml"`
	PackageJson   string `json:"package_json"`
}

// McpPull fetches an MCP server's source files from Aerostack.
// slug may be scoped (@username/mcp-name) or bare (mcp-name — API prefixes with caller's username).
func McpPull(apiKey, slug string) (*McpPullResponse, error) {
	url := fmt.Sprintf("%s/api/v1/cli/mcp/%s", getBaseURL(), slug)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Key", apiKey)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("MCP server '%s' not found", slug)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("mcp pull failed (%d): %s", resp.StatusCode, string(body))
	}

	var out McpPullResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &out, nil
}

// ─── Skill Pull ───────────────────────────────────────────────────────────────

// SkillPullResponse holds the files returned by GET /api/v1/cli/skill/:slug.
type SkillPullResponse struct {
	Slug    string `json:"slug"`
	Name    string `json:"name"`
	Content string `json:"content"`
}

// SkillPull fetches a skill's SKILL.md content from Aerostack.
// slug may be scoped (@username/skill-name) or bare (skill-name — API prefixes with caller's username).
func SkillPull(apiKey, slug string) (*SkillPullResponse, error) {
	url := fmt.Sprintf("%s/api/v1/cli/skill/%s", getBaseURL(), slug)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Key", apiKey)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("skill '%s' not found", slug)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("skill pull failed (%d): %s", resp.StatusCode, string(body))
	}

	var out SkillPullResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &out, nil
}

// TeamCheckMembership checks if the authenticated caller is a member of ownerUsername's team.
func TeamCheckMembership(apiKey, ownerUsername string) (bool, error) {
	url := fmt.Sprintf("%s/api/community/team/membership/%s", getBaseURL(), ownerUsername)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("X-API-Key", apiKey)

	resp, err := httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return false, fmt.Errorf("team membership check failed (%d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		IsMember bool `json:"isMember"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return false, err
	}
	return result.IsMember, nil
}
