package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestErrorUI_ShowsOptions(t *testing.T) {
	m := Model{errActive: true, errMsg: "boom"}
	v := m.View()
	if !strings.Contains(v, "Error: boom") {
		t.Fatalf("view missing error message; got: %q", v)
	}
	if !strings.Contains(v, "[r]etry") || !strings.Contains(v, "[s]kip") || !strings.Contains(v, "[a]bort") {
		t.Fatalf("view missing options; got: %q", v)
	}
}

func TestErrorUI_ChooseRetry(t *testing.T) {
	m := Model{errActive: true, errMsg: "boom"}
	p := tea.NewProgram(m, tea.WithoutRenderer(), tea.WithInput(strings.NewReader("r")))
	final, err := p.Run()
	if err != nil {
		t.Fatalf("program run failed: %v", err)
	}
	fm, ok := final.(Model)
	if !ok {
		t.Fatalf("final model type mismatch")
	}
	if fm.errChoice != choiceRetry {
		t.Fatalf("expected retry choice, got %v", fm.errChoice)
	}
}

func TestErrorUI_ChooseSkip(t *testing.T) {
	m := Model{errActive: true, errMsg: "boom"}
	p := tea.NewProgram(m, tea.WithoutRenderer(), tea.WithInput(strings.NewReader("s")))
	final, err := p.Run()
	if err != nil {
		t.Fatalf("program run failed: %v", err)
	}
	fm := final.(Model)
	if fm.errChoice != choiceSkip {
		t.Fatalf("expected skip choice, got %v", fm.errChoice)
	}
}

func TestErrorUI_ChooseAbort(t *testing.T) {
	m := Model{errActive: true, errMsg: "boom"}
	p := tea.NewProgram(m, tea.WithoutRenderer(), tea.WithInput(strings.NewReader("a")))
	final, err := p.Run()
	if err != nil {
		t.Fatalf("program run failed: %v", err)
	}
	fm := final.(Model)
	if fm.errChoice != choiceAbort {
		t.Fatalf("expected abort choice, got %v", fm.errChoice)
	}
}
