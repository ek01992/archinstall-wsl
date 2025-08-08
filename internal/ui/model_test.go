package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewModelViewContainsPlaceholder(t *testing.T) {
	m := New()
	v := m.View()
	if !strings.Contains(v, "ArchWSL TUI Configurator") {
		t.Fatalf("view missing app title; got: %q", v)
	}
	if !strings.Contains(v, "placeholder") {
		t.Fatalf("view missing placeholder text; got: %q", v)
	}
}

func TestProgramRunsAndQuits(t *testing.T) {
	m := New()
	p := tea.NewProgram(m, tea.WithoutRenderer(), tea.WithInput(strings.NewReader("q")))
	if _, err := p.Run(); err != nil {
		t.Fatalf("program run failed: %v", err)
	}
}
