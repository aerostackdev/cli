package commands

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const docsURL = "https://aerostack.dev/docs/cli"

// RenderWelcome prints a branded welcome screen to stdout (name, tagline, version, quick start, link).
func RenderWelcome(version string) {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")) // blue
	taglineStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))
	versionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("86"))
	sectionStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("252"))
	cmdStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("99"))
	linkStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Underline(true)

	lines := []string{
		"",
		titleStyle.Render("  ⚡ Aerostack"),
		taglineStyle.Render("  Build and deploy serverless applications on the edge."),
		"",
		versionStyle.Render("  " + version),
		"",
		sectionStyle.Render("  Quick start:"),
		cmdStyle.Render("    aerostack init [dir]   Scaffold a new project"),
		cmdStyle.Render("    aerostack dev          Start local development"),
		cmdStyle.Render("    aerostack deploy       Deploy to Cloudflare"),
		"",
		"  " + linkStyle.Render(docsURL),
		"",
		taglineStyle.Render("  Run aerostack --help for all commands."),
		"",
	}
	fmt.Println(strings.Join(lines, "\n"))
}
