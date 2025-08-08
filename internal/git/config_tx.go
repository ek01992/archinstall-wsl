package git

import (
	"strings"

	"archwsl-tui-configurator/internal/tx"
)

// configureGitTx sets git config with rollback to prior values on failure.
func configureGitTx(userName, userEmail string) (err error) {
	tr := tx.New()
	defer func() { if err != nil { _ = tr.Rollback() } }()

	prevName, _ := runCommandCapture("git", "config", "--global", "--get", "user.name")
	prevEmail, _ := runCommandCapture("git", "config", "--global", "--get", "user.email")
	prevName = strings.TrimSpace(prevName)
	prevEmail = strings.TrimSpace(prevEmail)

	tr.Defer(func() error {
		if prevName == "" {
			_ = runCommand("git", "config", "--global", "--unset", "user.name")
		} else {
			_ = runCommand("git", "config", "--global", "user.name", prevName)
		}
		if prevEmail == "" {
			_ = runCommand("git", "config", "--global", "--unset", "user.email")
		} else {
			_ = runCommand("git", "config", "--global", "user.email", prevEmail)
		}
		return nil
	})

	if err = configureGit(userName, userEmail); err != nil {
		return err
	}
	tr.Commit()
	return nil
}
