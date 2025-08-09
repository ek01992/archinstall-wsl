package app

import (
	tea "github.com/charmbracelet/bubbletea"

	"archwsl-tui-configurator/internal/ui"
)

type App struct{}

func New() *App { return &App{} }

// seams for program/model construction and execution
var (
	newModel   = func() tea.Model { return ui.NewWithActions(newAppActions()) }
	newProgram = func(m tea.Model, opts ...tea.ProgramOption) *tea.Program { return tea.NewProgram(m, opts...) }
	runProgram = func(p *tea.Program) (tea.Model, error) { return p.Run() }
)

func (a *App) Run() error {
	p := newProgram(newModel(), tea.WithAltScreen())
	_, err := runProgram(p)
	return err
}
