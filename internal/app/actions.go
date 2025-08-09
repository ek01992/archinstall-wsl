package app

import (
	"fmt"
	"strings"

	"archwsl-tui-configurator/internal/ui"
)

type appActions struct {
	p *Provider
}

func newAppActions() ui.Actions {
	return &appActions{p: NewProvider()}
}

func (a *appActions) Items() []ui.Action {
	return []ui.Action{
		{ID: "sysinfo", Title: "Show system info"},
		{ID: "nerdfont", Title: "Detect Nerd Font (Windows host)"},
		{ID: "load-config", Title: "Load saved config (preview)"},
	}
}

func (a *appActions) Execute(id string) (string, error) {
	switch id {
	case "sysinfo":
		var b strings.Builder
		b.WriteString("Show system info:\n")
		b.WriteString("WSL: ")
		if a.p.Platform.IsWSL() {
			b.WriteString("true\n")
		} else {
			b.WriteString("false\n")
		}
		b.WriteString("Host files editable (/mnt/c): ")
		if a.p.Platform.CanEditHostFiles() {
			b.WriteString("true\n")
		} else {
			b.WriteString("false\n")
		}
		if v := strings.TrimSpace(a.p.Platform.Getenv("WSL_DISTRO_NAME")); v != "" {
			b.WriteString("WSL_DISTRO_NAME: ")
			b.WriteString(v)
			b.WriteString("\n")
		}
		return b.String(), nil

	case "nerdfont":
		if a.p.NerdFont.Detect() {
			return "Detect Nerd Font (Windows host): detected", nil
		}
		return "Detect Nerd Font (Windows host): not detected", nil

	case "load-config":
		cfg := a.p.Config.Load("")
		var b strings.Builder
		b.WriteString("Load saved config (preview): ")
		if cfg.IsZero() {
			return "No saved config found.", nil
		}
		b.WriteString("\n")
		if cfg.Username != "" {
			b.WriteString("  Username: ")
			b.WriteString(cfg.Username)
			b.WriteString("\n")
		}
		if cfg.GitName != "" {
			b.WriteString("  GitName: ")
			b.WriteString(cfg.GitName)
			b.WriteString("\n")
		}
		if cfg.GitEmail != "" {
			b.WriteString("  GitEmail: ")
			b.WriteString(cfg.GitEmail)
			b.WriteString("\n")
		}
		if cfg.OhMyZshTheme != "" {
			b.WriteString("  OhMyZshTheme: ")
			b.WriteString(cfg.OhMyZshTheme)
			b.WriteString("\n")
		}
		if len(cfg.OhMyZshPlugins) > 0 {
			b.WriteString("  OhMyZshPlugins: ")
			b.WriteString(strings.Join(cfg.OhMyZshPlugins, ", "))
			b.WriteString("\n")
		}
		if cfg.DotfilesRepo != "" {
			b.WriteString("  DotfilesRepo: ")
			b.WriteString(cfg.DotfilesRepo)
			b.WriteString("\n")
		}
		b.WriteString("  NonInteractive: ")
		if cfg.NonInteractive {
			b.WriteString("true")
		} else {
			b.WriteString("false")
		}
		return b.String(), nil

	default:
		return "", fmt.Errorf("unknown action: %s", id)
	}
}
