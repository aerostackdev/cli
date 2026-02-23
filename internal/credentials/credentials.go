package credentials

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
)

const credentialsDir = ".aerostack"
const credentialsFile = "credentials.json"

type Credentials struct {
	APIKey string `json:"api_key"`
}

func credentialsPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, credentialsDir, credentialsFile), nil
}

func Load() (*Credentials, error) {
	path, err := credentialsPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var cred Credentials
	if err := json.Unmarshal(data, &cred); err != nil {
		return nil, err
	}
	if cred.APIKey == "" {
		return nil, nil
	}
	return &cred, nil
}

func Save(apiKey string) error {
	path, err := credentialsPath()
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	cred := Credentials{APIKey: apiKey}
	data, err := json.MarshalIndent(cred, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func Clear() error {
	path, err := credentialsPath()
	if err != nil {
		return err
	}
	return os.Remove(path)
}

// GetAPIKey returns the API key from environment, aerostack.toml, or credentials file.
func GetAPIKey() string {
	if k := os.Getenv("AEROSTACK_API_KEY"); k != "" {
		return k
	}

	// Try aerostack.toml in current directory
	if data, err := os.ReadFile("aerostack.toml"); err == nil {
		re := regexp.MustCompile(`(?m)^\s*api_key\s*=\s*"([^"]+)"`)
		matches := re.FindStringSubmatch(string(data))
		if len(matches) > 1 {
			return matches[1]
		}
	}

	// Fallback to home credentials
	cred, _ := Load()
	if cred != nil {
		return cred.APIKey
	}

	return ""
}
