package link

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const projectDir = ".aerostack"
const projectFile = "project.json"

type ProjectLink struct {
	ProjectID string `json:"project_id"`
}

func projectPath() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(wd, projectDir, projectFile), nil
}

func Load() (*ProjectLink, error) {
	path, err := projectPath()
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
	var link ProjectLink
	if err := json.Unmarshal(data, &link); err != nil {
		return nil, err
	}
	if link.ProjectID == "" {
		return nil, nil
	}
	return &link, nil
}

func Save(projectID string) error {
	path, err := projectPath()
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	link := ProjectLink{ProjectID: projectID}
	data, err := json.MarshalIndent(link, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
