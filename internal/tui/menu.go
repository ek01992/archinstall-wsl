package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// MenuModel represents the main menu state
type MenuModel struct {
	choices  []string
	cursor   int
	selected map[int]struct{}
}

// NewMenuModel creates a new menu model
func NewMenuModel() MenuModel {
	return MenuModel{
		choices: []string{
			"📦 Package Selection",
			"🌐 Network Configuration",
			"👤 User Management",
			"🔧 System Settings",
			"💾 Storage Configuration",
			"🔒 Security Settings",
			"🎨 Desktop Environment",
			"📋 Review Configuration",
			"🚀 Begin Installation",
		},
		selected: make(map[int]struct{}),
	}
}

// Init initializes the menu model
func (m MenuModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the menu view
func (m MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		case "enter", " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}

		case "r":
			// Reset selection
			m.selected = make(map[int]struct{})

		case "b":
			// Go back to welcome
			return m, func() tea.Msg {
				return SwitchViewMsg{View: WelcomeView}
			}
		}
	}

	return m, nil
}

// View renders the main menu
func (m MenuModel) View() string {
	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4")).
		PaddingBottom(1).
		PaddingLeft(2)

	header := headerStyle.Render("🏛️  ArchInstall WSL Configuration Menu")

	// Menu items
	menuStyle := lipgloss.NewStyle().
		PaddingLeft(2).
		PaddingTop(1)

	itemStyle := lipgloss.NewStyle().
		PaddingLeft(2).
		PaddingRight(2)

	selectedItemStyle := itemStyle.Copy().
		Foreground(lipgloss.Color("#EE6FF8")).
		Bold(true)

	completedItemStyle := itemStyle.Copy().
		Foreground(lipgloss.Color("#04B575")).
		Strikethrough(true)

	cursorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF75B7"))

	var menuItems []string
	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = cursorStyle.Render("→")
		}

		checked := " "
		if _, ok := m.selected[i]; ok {
			checked = "✓"
		}

		style := itemStyle
		if m.cursor == i {
			style = selectedItemStyle
		} else if _, ok := m.selected[i]; ok {
			style = completedItemStyle
		}

		item := style.Render(cursor + " [" + checked + "] " + choice)
		menuItems = append(menuItems, item)
	}

	menu := menuStyle.Render(lipgloss.JoinVertical(lipgloss.Left, menuItems...))

	// Instructions
	instructionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		PaddingTop(2).
		PaddingLeft(2)

	instructions := instructionStyle.Render(
		"↑/k: up • ↓/j: down • enter/space: select • r: reset • b: back • q/ctrl+c: quit",
	)

	// Container
	containerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#874BFD")).
		Padding(1).
		MarginTop(1).
		MarginLeft(1)

	content := lipgloss.JoinVertical(lipgloss.Left,
		header,
		menu,
		instructions,
	)

	return containerStyle.Render(content) + "\n"
}
