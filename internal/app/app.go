package app

import (
	tea "github.com/charmbracelet/bubbletea"

	"archwsl-tui-configurator/internal/ui"
)

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Run() error {
	p := tea.NewProgram(ui.New(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
