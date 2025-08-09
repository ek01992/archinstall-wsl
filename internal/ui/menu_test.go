package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

type fxActs struct{ items []Action }

func (f fxActs) Items() []Action                   { return f.items }
func (f fxActs) Execute(id string) (string, error) { return "ok", nil }

func TestMenu_NavigationBounds(t *testing.T) {
	m := NewWithActions(fxActs{items: []Action{
		{ID: "a", Title: "A"},
		{ID: "b", Title: "B"},
	}})
	var cmd tea.Cmd

	// down beyond end (should clamp at last)
	m, cmd = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	mc := m.(Model)
	if mc.idx != 1 {
		t.Fatalf("down clamp failed, idx=%d", mc.idx)
	}

	// up beyond start (should clamp at 0)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	mc = m.(Model)
	if mc.idx != 0 {
		t.Fatalf("up clamp failed, idx=%d", mc.idx)
	}
	_ = cmd // silence
}

func TestMenu_EnterDisabledWhileRunning(t *testing.T) {
	m := NewWithActions(fxActs{items: []Action{{ID: "a", Title: "A"}}})

	// First enter should schedule a command and set running=true
	var cmd1 tea.Cmd
	m, cmd1 = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd1 == nil {
		t.Fatalf("expected first enter to schedule command")
	}
	if !m.(Model).running {
		t.Fatalf("expected running=true after first enter")
	}

	// Second enter while running should not schedule another command
	var cmd2 tea.Cmd
	m, cmd2 = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd2 != nil {
		t.Fatalf("expected no command on second enter while running")
	}
}

func TestMenu_ViewHintsPresent(t *testing.T) {
	m := NewWithActions(fxActs{items: []Action{{ID: "a", Title: "A"}}})
	v := m.View()
	if !strings.Contains(v, "Select an action (↑/↓ or j/k). Press Enter to run, q to quit.") {
		t.Fatalf("missing menu hint: %q", v)
	}
}

func TestMenu_ActionErrorClearsLastOutput(t *testing.T) {
	// Start with a concrete model to set lastOutput
	mc := NewWithActions(fxActs{items: []Action{{ID: "a", Title: "A"}}}).(Model)
	mc.lastOutput = "ok"
	m, _ := mc.Update(actionDoneMsg{output: "", err: errStr("boom")})
	got := m.(Model)
	if got.lastOutput != "" {
		t.Fatalf("expected lastOutput cleared on error, got %q", got.lastOutput)
	}
	if got.lastErr != "boom" {
		t.Fatalf("expected lastErr=boom, got %q", got.lastErr)
	}
}

type errStr string

func (e errStr) Error() string { return string(e) }
