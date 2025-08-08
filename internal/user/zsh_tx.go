package user

import "archwsl-tui-configurator/internal/tx"

// installZshTx sets zsh and restores previous shell on failure.
func installZshTx() (err error) {
	tr := tx.New()
	defer func() { if err != nil { _ = tr.Rollback() } }()

	username := getTargetUsername()
	prev := getDefaultShell(username)
	tr.Defer(func() error { return runCommand("chsh", "-s", prev, username) })

	if err = installZsh(); err != nil {
		return err
	}
	tr.Commit()
	return nil
}
