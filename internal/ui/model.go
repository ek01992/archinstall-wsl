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

type Action struct {
	ID    string
	Title string
}

type Actions interface {
	Items() []Action
	Execute(id string) (string, error)
}

type actionDoneMsg struct {
	output string
	err    error
}

type Model struct {
	// Error UI state
	errActive bool
	errMsg    string
	errChoice errorChoice

	// Menu/action state
	acts       Actions
	items      []Action
	idx        int
	running    bool
	lastOutput string
	lastErr    string
}

func New() tea.Model {
	return Model{}
}

func NewWithActions(acts Actions) tea.Model {
	m := Model{acts: acts}
	if acts != nil {
		m.items = acts.Items()
	}
	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Error flow handling
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

		// Normal menu handling
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "up", "k":
			if m.idx > 0 {
				m.idx--
			}
		case "down", "j":
			if m.idx+1 < len(m.items) {
				m.idx++
			}
		case "enter":
			if m.acts != nil && len(m.items) > 0 && !m.running {
				sel := m.items[m.idx]
				id := sel.ID
				m.running = true
				m.lastErr = ""
				m.lastOutput = "Running " + strings.TrimSpace(sel.Title) + "..."
				return m, func() tea.Msg {
					out, err := m.acts.Execute(id)
					return actionDoneMsg{output: out, err: err}
				}
			}
		}

	case actionDoneMsg:
		m.running = false
		if msg.err != nil {
			m.lastErr = strings.TrimSpace(msg.err.Error())
			if m.lastErr == "" {
				m.lastErr = "error"
			}
			m.lastOutput = "" // clear stale success text on error
		} else {
			m.lastOutput = strings.TrimSpace(msg.output)
		}
		return m, nil
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
	b.WriteString("ArchWSL TUI Configurator\n\n")

	if len(m.items) > 0 {
		b.WriteString("Select an action (↑/↓ or j/k). Press Enter to run, q to quit.\n\n")
		for i, it := range m.items {
			prefix := "  "
			if i == m.idx {
				prefix = "> "
			}
			b.WriteString(prefix)
			b.WriteString(it.Title)
			b.WriteString("\n")
		}
		b.WriteString("\n")
		if strings.TrimSpace(m.lastErr) != "" {
			b.WriteString("Error: ")
			b.WriteString(m.lastErr)
			b.WriteString("\n")
		} else if strings.TrimSpace(m.lastOutput) != "" {
			b.WriteString(m.lastOutput)
			b.WriteString("\n")
		}
	} else {
		b.WriteString("Provision an idempotent, repeatable Arch Linux environment on WSL2.\n")
		b.WriteString("Safe to re-run anytime; changes only when needed.\n\n")
		b.WriteString("Press q to quit.")
		return b.String()
	}

	b.WriteString("Press q to quit.")
	return b.String()
}
