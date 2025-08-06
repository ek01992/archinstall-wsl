package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// WelcomeModel represents the welcome screen state
type WelcomeModel struct {
	showingWelcome bool
	timer          *time.Timer
}

// NewWelcomeModel creates a new welcome model
func NewWelcomeModel() WelcomeModel {
	return WelcomeModel{
		showingWelcome: true,
	}
}

// Init initializes the welcome model
func (m WelcomeModel) Init() tea.Cmd {
	// Auto-transition to menu after 3 seconds
	return tea.Tick(time.Second*3, func(t time.Time) tea.Msg {
		return SwitchViewMsg{View: MenuView}
	})
}

// Update handles messages for the welcome view
func (m WelcomeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter", " ":
			// Allow user to skip welcome screen
			return m, func() tea.Msg {
				return SwitchViewMsg{View: MenuView}
			}
		}
	case SwitchViewMsg:
		if msg.View == MenuView {
			m.showingWelcome = false
		}
	}

	return m, nil
}

// View renders the welcome screen
func (m WelcomeModel) View() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4")).
		PaddingTop(2).
		PaddingLeft(4)

	subtitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		PaddingLeft(4).
		PaddingBottom(2)

	instructionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#04B575")).
		PaddingLeft(4).
		PaddingTop(1)

	containerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#874BFD")).
		Padding(1, 2).
		MarginTop(2).
		MarginLeft(2)

	title := titleStyle.Render("🏛️  ArchInstall WSL TUI Configurator")
	subtitle := subtitleStyle.Render("A modern TUI for configuring ArchLinux installations on WSL")
	instruction := instructionStyle.Render("Press ENTER to continue or wait 3 seconds...")

	content := lipgloss.JoinVertical(lipgloss.Left,
		title,
		subtitle,
		instruction,
	)

	return containerStyle.Render(content) + "\n\n"
}
