package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/aerostackdev/cli/internal/api"
	"github.com/aerostackdev/cli/internal/credentials"
	"github.com/aerostackdev/cli/internal/link"
	"github.com/aerostackdev/cli/internal/printer"
	"github.com/spf13/cobra"
)

// NewLinkCommand creates the 'aerostack link' command
func NewLinkCommand() *cobra.Command {
	var writeToml bool

	cmd := &cobra.Command{
		Use:   "link [project-id-or-slug]",
		Short: "Link this directory to an Aerostack project",
		Long: `Link your local project to an Aerostack project.

Without arguments, lists your projects and lets you choose one interactively.
With an argument, links directly to the given project ID or slug.

Use --write-toml to also write project_id into aerostack.toml (persists across machines).`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				return linkDirect(args[0], writeToml)
			}
			return linkInteractive(writeToml)
		},
	}

	cmd.Flags().BoolVar(&writeToml, "write-toml", false, "Write project_id into aerostack.toml")

	return cmd
}

// linkDirect links to a project by ID or slug directly.
func linkDirect(idOrSlug string, writeToml bool) error {
	if err := link.Save(idOrSlug); err != nil {
		return fmt.Errorf("save link: %w", err)
	}
	printer.Success("Linked to project: %s", idOrSlug)

	if writeToml {
		if err := writeProjectIDToToml(idOrSlug); err != nil {
			printer.Warn("Could not write project_id to aerostack.toml: %v", err)
		} else {
			printer.Hint("Written project_id to aerostack.toml")
		}
	}

	fmt.Println()
	printer.Hint("Run 'aerostack deploy' to deploy to this project.")
	return nil
}

// linkInteractive lists the user's projects and prompts them to choose one.
func linkInteractive(writeToml bool) error {
	cred, _ := credentials.Load()
	if cred == nil || cred.APIKey == "" {
		return fmt.Errorf("not logged in. Run 'aerostack login' first.")
	}

	validateResp, err := api.Validate(cred.APIKey)
	if err != nil {
		return fmt.Errorf("API key invalid: %w", err)
	}
	if validateResp.KeyType != "account" {
		return fmt.Errorf("interactive link requires an account key. Use 'aerostack link <project-id>' with a project key.")
	}

	printer.Step("Fetching your projects...")
	projects, err := api.ListProjects(cred.APIKey)
	if err != nil {
		return fmt.Errorf("failed to list projects: %w", err)
	}

	if len(projects) == 0 {
		printer.Warn("No projects found. Create one at https://app.aerostack.dev or run 'aerostack deploy' to auto-create.")
		return nil
	}

	fmt.Println()
	fmt.Println("Your projects:")
	fmt.Println()
	for i, p := range projects {
		fmt.Printf("  [%d] %s\n", i+1, p.Name)
		fmt.Printf("      ID: %s | URL: %s\n", p.ID, p.URL)
	}
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter number to link (or press Enter to cancel): ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		printer.Hint("Cancelled.")
		return nil
	}

	var idx int
	if _, err := fmt.Sscanf(input, "%d", &idx); err != nil || idx < 1 || idx > len(projects) {
		return fmt.Errorf("invalid selection: %q", input)
	}

	chosen := projects[idx-1]
	if err := link.Save(chosen.ID); err != nil {
		return fmt.Errorf("save link: %w", err)
	}
	printer.Success("Linked to project: %s (%s)", chosen.Name, chosen.ID)

	if writeToml {
		if err := writeProjectIDToToml(chosen.ID); err != nil {
			printer.Warn("Could not write project_id to aerostack.toml: %v", err)
		} else {
			printer.Hint("Written project_id to aerostack.toml")
		}
	}

	fmt.Println()
	printer.Hint("Run 'aerostack deploy' to deploy to this project.")
	return nil
}

// writeProjectIDToToml upserts project_id = "..." in aerostack.toml.
func writeProjectIDToToml(projectID string) error {
	const filename = "aerostack.toml"
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	content := string(data)
	line := fmt.Sprintf("project_id = \"%s\"", projectID)

	// Replace existing project_id line if present
	lines := strings.Split(content, "\n")
	replaced := false
	for i, l := range lines {
		trimmed := strings.TrimSpace(l)
		if strings.HasPrefix(trimmed, "project_id") {
			lines[i] = line
			replaced = true
			break
		}
	}

	if !replaced {
		// Insert after the name = "..." line
		newLines := []string{}
		for _, l := range lines {
			newLines = append(newLines, l)
			if strings.HasPrefix(strings.TrimSpace(l), "name") && strings.Contains(l, "=") {
				newLines = append(newLines, line)
			}
		}
		lines = newLines
	}

	return os.WriteFile(filename, []byte(strings.Join(lines, "\n")), 0644)
}
