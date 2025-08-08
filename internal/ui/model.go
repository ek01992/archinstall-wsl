package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type errorChoice int

const (
	choiceNone errorChoice = iota
	choiceRetry
	choiceSkip
	choiceAbort
)

type Model struct {
	errActive bool
	errMsg    string
	errChoice errorChoice
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
		if m.errActive {
			switch msg.String() {
			case "r":
				m.errChoice = choiceRetry
				return m, tea.Quit
			case "s":
				m.errChoice = choiceSkip
				return m, tea.Quit
			case "a":
				m.errChoice = choiceAbort
				return m, tea.Quit
			case "ctrl+c", "q", "esc":
				m.errChoice = choiceAbort
				return m, tea.Quit
			}
		}
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) View() string {
	if m.errActive {
		var b strings.Builder
		b.WriteString("ArchWSL TUI Configurator\n\n")
		b.WriteString("Error: ")
		b.WriteString(strings.TrimSpace(m.errMsg))
		b.WriteString("\n\n")
		b.WriteString("[r]etry  [s]kip  [a]bort\n")
		return b.String()
	}
	var b strings.Builder
	b.WriteString("ArchWSL TUI Configurator\n")
	b.WriteString("\n")
	b.WriteString("Provision an idempotent, repeatable Arch Linux environment on WSL2.\n")
	b.WriteString("Safe to re-run anytime; changes only when needed.\n")
	b.WriteString("\n")
	b.WriteString("Press q to quit.")
	return b.String()
}
