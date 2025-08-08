package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewModelViewContainsTitleAndDescription(t *testing.T) {
	m := New()
	v := m.View()
	if !strings.Contains(v, "ArchWSL TUI Configurator") {
		t.Fatalf("view missing app title; got: %q", v)
	}
	if !strings.Contains(v, "idempotent") {
		t.Fatalf("view missing idempotent description; got: %q", v)
	}
	if !strings.Contains(v, "Press q to quit") {
		t.Fatalf("view missing quit hint; got: %q", v)
	}
}

func TestProgramRunsAndQuits(t *testing.T) {
	m := New()
	p := tea.NewProgram(m, tea.WithoutRenderer(), tea.WithInput(strings.NewReader("q")))
	if _, err := p.Run(); err != nil {
		t.Fatalf("program run failed: %v", err)
	}
}
