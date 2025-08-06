package tui

// ViewType represents the different views in the application
type ViewType int

const (
	WelcomeView ViewType = iota
	MenuView
)

// SwitchViewMsg is a message to switch between views
type SwitchViewMsg struct {
	View ViewType
}
