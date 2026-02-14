package neon

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const neonAPIBase = "https://console.neon.tech/api/v2"

// Client is a Neon API client
type Client struct {
	apiKey string
	client *http.Client
}

// NewClient creates a new Neon API client
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		client: &http.Client{},
	}
}

// Project represents a Neon project
type Project struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	RegionID  string `json:"region_id"`
	CreatedAt string `json:"created_at"`
}

// Database represents a Neon database
type Database struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	OwnerName string `json:"owner_name"`
}

// ConnectionDetails holds DB connection information
type ConnectionDetails struct {
	Host         string `json:"host"`
	DatabaseName string `json:"database_name"`
	User         string `json:"user"`
	Password     string `json:"password"`
	Port         int    `json:"port"`
}

// CreateProjectRequest is the request body for creating a project
type CreateProjectRequest struct {
	Project struct {
		Name     string `json:"name,omitempty"`
		RegionID string `json:"region_id,omitempty"`
	} `json:"project"`
}

// CreateProjectResponse is the response from creating a project.
// Neon API v2 returns connection_uri as a string, or connection_uris as an array.
// Branches may also contain connection_uris.
type CreateProjectResponse struct {
	Project        Project `json:"project"`
	ConnectionURI  string  `json:"connection_uri"` // Top-level string (if present)
	ConnectionURIs []struct {
		ConnectionURI string `json:"connection_uri"`
	} `json:"connection_uris"`
	Branches []struct {
		ConnectionURIs []struct {
			ConnectionURI string `json:"connection_uri"`
		} `json:"connection_uris"`
	} `json:"branches"`
	// Legacy: parsed connection details (used if URI not in response)
	Connection ConnectionDetails `json:"connection"`
	Password   string            `json:"password"`
}

// GetConnectionString returns the first available connection string from the response.
func (r *CreateProjectResponse) GetConnectionString() string {
	if r.ConnectionURI != "" {
		return r.ConnectionURI
	}
	if len(r.ConnectionURIs) > 0 && r.ConnectionURIs[0].ConnectionURI != "" {
		return r.ConnectionURIs[0].ConnectionURI
	}
	for _, b := range r.Branches {
		for _, cu := range b.ConnectionURIs {
			if cu.ConnectionURI != "" {
				return cu.ConnectionURI
			}
		}
	}
	// Fallback: build from Connection + Password (legacy/alternate API shape)
	if r.Connection.Host != "" && r.Password != "" {
		return BuildConnectionString(r.Connection, r.Password)
	}
	return ""
}

// CreateProject creates a new Neon project
func (c *Client) CreateProject(name, region string) (*CreateProjectResponse, error) {
	req := CreateProjectRequest{}
	req.Project.Name = name
	if region != "" {
		req.Project.RegionID = region
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", neonAPIBase+"/projects", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	var result CreateProjectResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// BuildConnectionString builds a Postgres connection string
func BuildConnectionString(details ConnectionDetails, password string) string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s",
		details.User, password, details.Host, details.Port, details.DatabaseName)
}

// GetAPIKeyFromEnv retrieves the Neon API key from environment variables
func GetAPIKeyFromEnv() (string, error) {
	apiKey := os.Getenv("NEON_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("NEON_API_KEY environment variable not set. Get your API key from https://console.neon.tech/app/settings/api-keys")
	}
	return apiKey, nil
}
