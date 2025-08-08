package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// fakeErrorModel simulates a simple error UI flow end-to-end without rendering.
// It starts with an active error and accepts r/s/a to quit with a choice.
type fakeErrorModel struct {
	choice string
}

func (m fakeErrorModel) Init() tea.Cmd { return nil }

func (m fakeErrorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "r":
			m.choice = "retry"
			return m, tea.Quit
		case "s":
			m.choice = "skip"
			return m, tea.Quit
		case "a", "q", "ctrl+c", "esc":
			m.choice = "abort"
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m fakeErrorModel) View() string {
	return "Error: boom\n[r]etry  [s]kip  [a]bort\n"
}

func TestApp_E2E_ErrorRetryFlow(t *testing.T) {
	// Save and restore seams
	origNewProgram := newProgram
	origNewModel := newModel
	origRunProgram := runProgram
	defer func() {
		newProgram = origNewProgram
		newModel = origNewModel
		runProgram = origRunProgram
	}()

	// Inject our fake model that shows an error
	newModel = func() tea.Model { return fakeErrorModel{} }

	// Ensure program runs headless and receives an "r" keypress
	newProgram = func(m tea.Model, opts ...tea.ProgramOption) *tea.Program {
		return tea.NewProgram(m, tea.WithoutRenderer(), tea.WithInput(strings.NewReader("r")))
	}

	var final tea.Model
	runProgram = func(p *tea.Program) (tea.Model, error) {
		fm, err := p.Run()
		final = fm
		return fm, err
	}

	application := New()
	if err := application.Run(); err != nil {
		t.Fatalf("app run failed: %v", err)
	}

	fm, ok := final.(fakeErrorModel)
	if !ok {
		t.Fatalf("final model type mismatch: %T", final)
	}
	if fm.choice != "retry" {
		t.Fatalf("expected retry choice, got %q", fm.choice)
	}
}
