package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	quitting bool
}

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
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) View() string {
	return "ArchWSL TUI Configurator\n\nThis is a placeholder view.\n\nPress q to quit."
}
