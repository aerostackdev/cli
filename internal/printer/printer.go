package printer

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors (Enterprise Palette)
	BrandCyan = lipgloss.Color("#00E5FF")
	Emerald   = lipgloss.Color("#00E676")
	Amber     = lipgloss.Color("#FFD740")
	Coral     = lipgloss.Color("#FF5252")
	Slate     = lipgloss.Color("#78909C")
	Lavender  = lipgloss.Color("#B39DDB")
	OffWhite  = lipgloss.Color("#ECEFF1")

	// Styles
	boldStyle    = lipgloss.NewStyle().Bold(true)
	brandStyle   = lipgloss.NewStyle().Foreground(BrandCyan)
	successStyle = lipgloss.NewStyle().Foreground(Emerald)
	warnStyle    = lipgloss.NewStyle().Foreground(Amber)
	errorStyle   = lipgloss.NewStyle().Foreground(Coral)
	mutedStyle   = lipgloss.NewStyle().Foreground(Slate)
	cmdStyle     = lipgloss.NewStyle().Foreground(Lavender)

	// Glyphs
	GlyphStep    = brandStyle.Render("◆")
	GlyphSuccess = successStyle.Render("✓")
	GlyphWarn    = warnStyle.Render("⚠")
	GlyphError   = errorStyle.Render("✗")
	GlyphHint    = mutedStyle.Render("→")
)

// Success prints a success message with a green checkmark.
func Success(msg string, args ...any) {
	formatted := fmt.Sprintf(msg, args...)
	fmt.Printf("%s %s\n", GlyphSuccess, successStyle.Render(formatted))
}

// Step prints a step header with a brand cyan diamond.
func Step(msg string, args ...any) {
	formatted := fmt.Sprintf(msg, args...)
	fmt.Printf("%s %s\n", GlyphStep, boldStyle.Foreground(OffWhite).Render(formatted))
}

// Hint prints a muted hinting message with an arrow.
func Hint(msg string, args ...any) {
	formatted := fmt.Sprintf(msg, args...)
	fmt.Printf("%s %s\n", GlyphHint, mutedStyle.Render(formatted))
}

// Warn prints a warning message in amber.
func Warn(msg string, args ...any) {
	formatted := fmt.Sprintf(msg, args...)
	fmt.Printf("%s %s\n", GlyphWarn, warnStyle.Render(formatted))
}

// Error prints an error message in coral.
func Error(msg string, args ...any) {
	formatted := fmt.Sprintf(msg, args...)
	fmt.Printf("%s %s\n", GlyphError, errorStyle.Render(formatted))
}

// Command formats an inline command string in lavender color.
func Command(cmd string) string {
	return cmdStyle.Render(cmd)
}

// Rule prints a full width muted horizontal rule.
func Rule() string {
	return mutedStyle.Render("──────────────────────────────────────────────────")
}

// Header prints a bold, brand-colored title block.
func Header(title string) {
	fmt.Println()
	fmt.Println(boldStyle.Foreground(BrandCyan).Render(fmt.Sprintf("  %s", title)))
	fmt.Println()
}

// KeyVal formats an aligned key-value pair for lists or summaries.
func KeyVal(key, val string) string {
	return fmt.Sprintf("  %s %s", mutedStyle.Render(fmt.Sprintf("%-12s", key+":")), val)
}

// Link returns an underlined cyan link.
func Link(url string) string {
	return brandStyle.Underline(true).Render(url)
}
