package selfheal

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	err      error
	proposal string
	diff     string
	quitting bool
	accepted bool
}

func initialModel(err error, proposal, diff string) model {
	return model{
		err:      err,
		proposal: proposal,
		diff:     diff,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "n":
			m.quitting = true
			return m, tea.Quit
		case "enter", "y":
			m.accepted = true
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.quitting {
		return ""
	}

	doc := strings.Builder{}

	// Header
	headerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	doc.WriteString(headerStyle.Render("🚨 Error Detected"))
	doc.WriteString("\n\n")
	doc.WriteString(fmt.Sprintf("%v\n\n", m.err))

	// Proposal
	doc.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Render("🤖 AI Proposal:"))
	doc.WriteString("\n")
	doc.WriteString(m.proposal)
	doc.WriteString("\n\n")

	// Diff (if any)
	if m.diff != "" {
		doc.WriteString(lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Render(m.diff))
		doc.WriteString("\n\n")
	}

	// Footer
	doc.WriteString(lipgloss.NewStyle().Faint(true).Render("Apply this fix? [y/N]"))

	return doc.String()
}

// ShowProposal renders the TUI and returns true if the user accepted the fix
func ShowProposal(originalErr error, proposal, diff string) (bool, error) {
	p := tea.NewProgram(initialModel(originalErr, proposal, diff))
	m, err := p.Run()
	if err != nil {
		return false, err
	}

	if finalModel, ok := m.(model); ok {
		return finalModel.accepted, nil
	}
	return false, nil
}
