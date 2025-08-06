package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewModel(t *testing.T) {
	model := NewModel()

	if model.currentView != WelcomeView {
		t.Errorf("Expected initial view to be WelcomeView, got %v", model.currentView)
	}

	if model.quitting {
		t.Error("Expected quitting to be false initially")
	}
}

func TestModelInit(t *testing.T) {
	model := NewModel()
	cmd := model.Init()

	if cmd == nil {
		t.Error("Expected Init to return a command")
	}
}

func TestModelUpdate(t *testing.T) {
	model := NewModel()

	// Test quit message
	updatedModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	m := updatedModel.(Model)

	if !m.quitting {
		t.Error("Expected quitting to be true after ctrl+c")
	}

	if cmd == nil {
		t.Error("Expected quit command to be returned")
	}
}

func TestWelcomeModel(t *testing.T) {
	welcome := NewWelcomeModel()

	if !welcome.showingWelcome {
		t.Error("Expected showingWelcome to be true initially")
	}

	// Test init returns a command
	cmd := welcome.Init()
	if cmd == nil {
		t.Error("Expected Init to return a command")
	}

	// Test view rendering
	view := welcome.View()
	if len(view) == 0 {
		t.Error("Expected view to render content")
	}
}

func TestMenuModel(t *testing.T) {
	menu := NewMenuModel()

	if len(menu.choices) == 0 {
		t.Error("Expected menu to have choices")
	}

	if menu.cursor != 0 {
		t.Error("Expected initial cursor position to be 0")
	}

	// Test view rendering
	view := menu.View()
	if len(view) == 0 {
		t.Error("Expected view to render content")
	}
}

func TestMenuNavigation(t *testing.T) {
	menu := NewMenuModel()
	initialCursor := menu.cursor

	// Test down movement
	updatedModel, _ := menu.Update(tea.KeyMsg{Type: tea.KeyDown})
	m := updatedModel.(MenuModel)

	if m.cursor != initialCursor+1 {
		t.Errorf("Expected cursor to move down from %d to %d, got %d",
			initialCursor, initialCursor+1, m.cursor)
	}

	// Test up movement
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updatedModel.(MenuModel)

	if m.cursor != initialCursor {
		t.Errorf("Expected cursor to move back to %d, got %d", initialCursor, m.cursor)
	}
}

func TestSwitchViewMessage(t *testing.T) {
	msg := SwitchViewMsg{View: MenuView}

	if msg.View != MenuView {
		t.Errorf("Expected view to be MenuView, got %v", msg.View)
	}
}
