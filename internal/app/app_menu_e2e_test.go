package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"archwsl-tui-configurator/internal/ui"
)

type fakeActions struct {
	calls int
}

func (f *fakeActions) Items() []ui.Action {
	return []ui.Action{{ID: "x", Title: "Test Action"}}
}

func (f *fakeActions) Execute(id string) (string, error) {
	f.calls++
	return "ok", nil
}

func TestApp_E2E_RunOneActionThenQuit(t *testing.T) {
	origNewModel := newModel
	origNewProgram := newProgram
	defer func() { newModel = origNewModel; newProgram = origNewProgram }()

	f := &fakeActions{}
	newModel = func() tea.Model { return ui.NewWithActions(f) }
	newProgram = func(m tea.Model, opts ...tea.ProgramOption) *tea.Program {
		// Enter to run first action, then q to quit
		return tea.NewProgram(m, tea.WithoutRenderer(), tea.WithInput(strings.NewReader("\nq")))
	}

	a := New()
	if err := a.Run(); err != nil {
		t.Fatalf("app run failed: %v", err)
	}
	if f.calls != 1 {
		t.Fatalf("expected 1 action call, got %d", f.calls)
	}
}
