package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

func getBaseURL() string {
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
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Key", apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
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

	resp, err := http.DefaultClient.Do(req)
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
}

type DeployError struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Details string `json:"details"`
	} `json:"error"`
}

func Deploy(apiKey string, workerPath string, env string, projectName string, isPublic bool, isPrivate bool) (*DeployResponse, error) {
	workerData, err := os.ReadFile(workerPath)
	if err != nil {
		return nil, fmt.Errorf("read worker file: %w", err)
	}

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	_ = w.WriteField("env", env)
	if projectName != "" {
		_ = w.WriteField("name", projectName)
	}
	// Add is_public field if explicitly set
	if isPublic {
		_ = w.WriteField("isPublic", "true")
	} else if isPrivate {
		_ = w.WriteField("isPublic", "false")
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

	url := getBaseURL() + "/api/v1/cli/deploy"
	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Key", apiKey)
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
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

	resp, err := http.DefaultClient.Do(req)
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
