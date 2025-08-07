package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// View represents the current view in the application
type View int

const (
	WelcomeView View = iota
	MainMenuView
)

// Model represents the main application model
type Model struct {
	view     View
	quitting bool
	width    int
	height   int
}

// InitialModel returns the initial model
func InitialModel() Model {
	return Model{
		view: WelcomeView,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "enter", " ":
			if m.view == WelcomeView {
				m.view = MainMenuView
				return m, nil
			}
		case "esc":
			if m.view == MainMenuView {
				m.view = WelcomeView
				return m, nil
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

// View renders the current view
func (m Model) View() string {
	if m.quitting {
		return "\n  Goodbye!\n\n"
	}

	switch m.view {
	case WelcomeView:
		return m.renderWelcomeView()
	case MainMenuView:
		return m.renderMainMenuView()
	default:
		return "Unknown view"
	}
}

// renderWelcomeView renders the welcome screen
func (m Model) renderWelcomeView() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00ff00")).
		Align(lipgloss.Center).
		Render("Welcome to archinstall-wsl")

	subtitle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Align(lipgloss.Center).
		Render("WSL Arch Linux Installer")

	instructions := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#cccccc")).
		Align(lipgloss.Center).
		Render("Press Enter or Space to continue...")

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		"",
		subtitle,
		"",
		instructions,
	)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}

// renderMainMenuView renders the main menu
func (m Model) renderMainMenuView() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00ff00")).
		Align(lipgloss.Center).
		Render("Main Menu")

	menuItems := []string{
		"1. Install Arch Linux",
		"2. Configure WSL",
		"3. System Information",
		"4. Exit",
	}

	menu := ""
	for _, item := range menuItems {
		menu += lipgloss.NewStyle().
			Foreground(lipgloss.Color("#cccccc")).
			Render(item) + "\n"
	}

	instructions := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Align(lipgloss.Center).
		Render("Press ESC to go back")

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		"",
		menu,
		"",
		instructions,
	)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}
