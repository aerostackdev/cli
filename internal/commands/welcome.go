package commands

import (
	"fmt"

	"github.com/aerostackdev/cli/internal/printer"
)

const docsURL = "https://aerostack.dev/docs/cli"

// RenderWelcome prints a branded welcome screen to stdout (name, tagline, version, quick start, link).
func RenderWelcome(version string) {
	printer.Header("Aerostack CLI")

	fmt.Println(printer.KeyVal("Version", "v"+version))
	fmt.Println()

	printer.Step("Quick Start")
	fmt.Println(printer.KeyVal(printer.Command("aerostack init"), "Scaffold a new project"))
	fmt.Println(printer.KeyVal(printer.Command("aerostack dev"), "Start local development"))
	fmt.Println(printer.KeyVal(printer.Command("aerostack deploy"), "Deploy to Cloudflare"))
	fmt.Println()

	printer.Hint("Docs: %s", printer.Link(docsURL))
	printer.Hint("Run 'aerostack --help' for all commands.")
	fmt.Println()
}
