package user

import (
	"fmt"

	"archwsl-tui-configurator/internal/tx"
)

// createUserTx wraps createUser with rollback semantics.
// It registers compensating actions for created user and sudoers changes.
func createUserTx(username, password string) (err error) {
	tr := tx.New()
	defer func() {
		if err != nil {
			_ = tr.Rollback()
		}
	}()

	userExisted := doesUserExist(username)
	if !userExisted {
		// Undo: delete user and home if we created it
		tr.Defer(func() error { return runCommand("userdel", "-r", username) })
	}

	// Sudoers file may be overwritten; register undo to restore previous content
	sudoersFile := sudoersDPath + "/010_wheel_nopasswd"
	prev, prevErr := readFile(sudoersFile)
	if prevErr == nil {
		tr.Defer(func() error { return writeFile(sudoersFile, prev, 0o440) })
	} else {
		tr.Defer(func() error { return runCommand("rm", "-f", sudoersFile) })
	}

	if err = createUser(username, password); err != nil {
		return fmt.Errorf("createUser failed: %w", err)
	}

	tr.Commit()
	return nil
}
