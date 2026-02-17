package devserver

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

const capnpTemplate = `
using Workerd = import "/workerd/workerd.capnp";

const config :Workerd.Config = (
  services = [
    (name = "main", worker = .mainWorker),
  ],
  sockets = [
    ( name = "http",
      address = "*:{{.Port}}",
      http = (),
      service = "main"
    ),
  ],
);

const mainWorker :Workerd.Worker = (
  modules = [
    (name = "main", esModule = embed "{{.BundlePath}}"),
  ],
  compatibilityDate = "{{.CompatibilityDate}}",
);
`

// ConfigData holds the data for the Cap'n Proto template
type ConfigData struct {
	Port              int
	BundlePath        string
	DBPath            string
	CompatibilityDate string
	BuildCommand      string
	DevCommand        string
	DeployCommand     string
}

// GenerateConfig creates a .capnp configuration file for workerd
func GenerateConfig(dotAerostack string, data ConfigData) (string, error) {
	configPath := filepath.Join(dotAerostack, "config.capnp")

	tmpl, err := template.New("capnp").Parse(capnpTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse capnp template: %w", err)
	}

	file, err := os.Create(configPath)
	if err != nil {
		return "", fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, data); err != nil {
		return "", fmt.Errorf("failed to execute config template: %w", err)
	}

	fmt.Printf("📄 workerd configuration generated: %s\n", configPath)
	return configPath, nil
}
