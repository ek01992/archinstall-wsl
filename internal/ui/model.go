package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct{}

func New() tea.Model {
	return Model{}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) View() string {
	var b strings.Builder
	b.WriteString("ArchWSL TUI Configurator\n")
	b.WriteString("\n")
	b.WriteString("Provision an idempotent, repeatable Arch Linux environment on WSL2.\n")
	b.WriteString("Safe to re-run anytime; changes only when needed.\n")
	b.WriteString("\n")
	b.WriteString("Press q to quit.")
	return b.String()
}
