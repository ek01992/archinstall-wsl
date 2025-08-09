package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestApp_E2E_ErrorAbortFlow(t *testing.T) {
	origNewProgram := newProgram
	origNewModel := newModel
	origRunProgram := runProgram
	t.Cleanup(func() { newProgram = origNewProgram; newModel = origNewModel; runProgram = origRunProgram })

	newModel = func() tea.Model { return fakeErrorModel{} }
	newProgram = func(m tea.Model, opts ...tea.ProgramOption) *tea.Program {
		return tea.NewProgram(m, tea.WithoutRenderer(), tea.WithInput(strings.NewReader("a")))
	}

	var final tea.Model
	runProgram = func(p *tea.Program) (tea.Model, error) {
		fm, err := p.Run()
		final = fm
		return fm, err
	}

	application := New()
	if err := application.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	fm, ok := final.(fakeErrorModel)
	if !ok {
		t.Fatalf("final model type mismatch: %T", final)
	}
	if fm.choice != "abort" {
		t.Fatalf("expected abort choice, got %q", fm.choice)
	}
}
