package user

import (
	"strings"

	"archwsl-tui-configurator/internal/tx"
)

// installZshTx sets zsh and restores previous shell on failure.
func installZshTx() (err error) {
	if defaultService != nil {
		return defaultService.InstallZshTx()
	}
	tr := tx.New()
	defer func() {
		if err != nil {
			_ = tr.Rollback()
		}
	}()

	username := getTargetUsername()
	prev := getDefaultShell(username)
	if strings.TrimSpace(prev) != "" {
		p := prev
		u := username
		tr.Defer(func() error { return runCommand("chsh", "-s", p, u) })
	}

	if err = installZsh(); err != nil {
		return err
	}
	tr.Commit()
	return nil
}
