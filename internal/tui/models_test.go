package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestInitialModel(t *testing.T) {
	model := InitialModel()

	if model.view != WelcomeView {
		t.Errorf("Expected initial view to be WelcomeView, got %v", model.view)
	}

	if model.quitting {
		t.Error("Expected initial model to not be quitting")
	}
}

func TestModelUpdate(t *testing.T) {
	model := InitialModel()

	t.Run("Enter key transitions from welcome to main menu", func(t *testing.T) {
		model.view = WelcomeView
		updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

		if updatedModel.(Model).view != MainMenuView {
			t.Error("Expected view to transition to MainMenuView")
		}
	})

	t.Run("Space key transitions from welcome to main menu", func(t *testing.T) {
		model.view = WelcomeView
		updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeySpace})

		if updatedModel.(Model).view != MainMenuView {
			t.Error("Expected view to transition to MainMenuView")
		}
	})

	t.Run("ESC key transitions from main menu to welcome", func(t *testing.T) {
		model.view = MainMenuView
		updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEscape})

		if updatedModel.(Model).view != WelcomeView {
			t.Error("Expected view to transition to WelcomeView")
		}
	})

	t.Run("Ctrl+C sets quitting flag", func(t *testing.T) {
		model.quitting = false
		updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

		if !updatedModel.(Model).quitting {
			t.Error("Expected quitting flag to be set")
		}
	})

	t.Run("Q key sets quitting flag", func(t *testing.T) {
		model.quitting = false
		updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

		if !updatedModel.(Model).quitting {
			t.Error("Expected quitting flag to be set")
		}
	})
}

func TestModelView(t *testing.T) {
	model := InitialModel()
	model.width = 80
	model.height = 24

	t.Run("Welcome view contains expected text", func(t *testing.T) {
		model.view = WelcomeView
		view := model.View()

		if !strings.Contains(view, "Welcome to archinstall-wsl") {
			t.Error("Expected welcome view to contain title")
		}

		if !strings.Contains(view, "WSL Arch Linux Installer") {
			t.Error("Expected welcome view to contain subtitle")
		}

		if !strings.Contains(view, "Press Enter or Space to continue") {
			t.Error("Expected welcome view to contain instructions")
		}
	})

	t.Run("Main menu view contains expected text", func(t *testing.T) {
		model.view = MainMenuView
		view := model.View()

		if !strings.Contains(view, "Main Menu") {
			t.Error("Expected main menu view to contain title")
		}

		if !strings.Contains(view, "Install Arch Linux") {
			t.Error("Expected main menu view to contain menu items")
		}

		if !strings.Contains(view, "Press ESC to go back") {
			t.Error("Expected main menu view to contain instructions")
		}
	})

	t.Run("Quitting view shows goodbye message", func(t *testing.T) {
		model.quitting = true
		view := model.View()

		if !strings.Contains(view, "Goodbye") {
			t.Error("Expected quitting view to contain goodbye message")
		}
	})
}
