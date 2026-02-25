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
	Timeout: 10 * time.Second,
}

func getBaseURL() string {
	return BaseURL()
}

// BaseURL returns the Aerostack API base URL (env AEROSTACK_API_URL or default).
func BaseURL() string {
	if u := os.Getenv("AEROSTACK_API_URL"); u != "" {
		return u
	}
	return "https://api.aerocall.ai"
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

// SendTelemetry pushes error logs to the Aerostack API for system learning.
func SendTelemetry(apiKey, projectID, errorLog string) error {
	if apiKey == "" {
		return fmt.Errorf("API key required for telemetry")
	}

	url := getBaseURL() + "/api/v1/cli/telemetry/errors"
	payload := map[string]string{
		"projectId": projectID,
		"logs":      errorLog,
		"os":        "mac", // Can be runtime.GOOS
	}

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
	ID             string      `json:"id"`
	Slug           string      `json:"slug"`
	Name           string      `json:"name"`
	Description    string      `json:"description"`
	Readme         string      `json:"readme"`
	Category       string      `json:"category"`
	Tags           []string    `json:"tags"`
	Language       string      `json:"language"`
	Runtime        string      `json:"runtime"`
	Code           string      `json:"code"`
	ConfigSchema   interface{} `json:"config_schema"`
	License        string      `json:"license"`
	Version        string      `json:"version"`
	Status         string      `json:"status"`
	AuthorUsername string      `json:"author_username"`
	URL            string      `json:"url"`
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

func Deploy(apiKey string, files map[string]string, env string, projectName string, isPublic bool, isPrivate bool, bindingsJSON string, compatDate string, compatFlags []string) (*DeployResponse, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	_ = w.WriteField("env", env)
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
