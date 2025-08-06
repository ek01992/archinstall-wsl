package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Model represents the main application state
type Model struct {
	currentView ViewType
	welcome     WelcomeModel
	menu        MenuModel
	quitting    bool
	width       int
	height      int
}

// NewModel creates a new main model
func NewModel() Model {
	return Model{
		currentView: WelcomeView,
		welcome:     NewWelcomeModel(),
		menu:        NewMenuModel(),
		quitting:    false,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	switch m.currentView {
	case WelcomeView:
		return m.welcome.Init()
	case MenuView:
		return m.menu.Init()
	default:
		return nil
	}
}

// Update handles messages and updates the model state
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case SwitchViewMsg:
		m.currentView = msg.View
		switch msg.View {
		case MenuView:
			return m, m.menu.Init()
		case WelcomeView:
			return m, m.welcome.Init()
		}
	}

	// Update the current view
	var cmd tea.Cmd
	switch m.currentView {
	case WelcomeView:
		var model tea.Model
		model, cmd = m.welcome.Update(msg)
		m.welcome = model.(WelcomeModel)

	case MenuView:
		var model tea.Model
		model, cmd = m.menu.Update(msg)
		m.menu = model.(MenuModel)
	}

	return m, cmd
}

// View renders the current view
func (m Model) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	switch m.currentView {
	case WelcomeView:
		return m.welcome.View()
	case MenuView:
		return m.menu.View()
	default:
		return "Unknown view\n"
	}
}
