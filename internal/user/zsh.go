package user

import (
	"fmt"
	stduser "os/user"
	"strings"

	"archwsl-tui-configurator/internal/logx"
)

var (
	// getTargetUsername is a seam that returns the user whose shell should be set to zsh.
	// By default it returns the current user.
	getTargetUsername = func() string {
		u, err := stduser.Current()
		if err != nil || u == nil {
			return ""
		}
		return u.Username
	}
)

const zshPath = "/usr/bin/zsh"

// installZsh sets the target user's login shell to zsh, idempotently.
// It first checks the current shell and performs no action if already zsh.
// It tries `chsh -s /usr/bin/zsh <user>` and falls back to `usermod -s /usr/bin/zsh <user>`.
func installZsh() error {
	username := strings.TrimSpace(getTargetUsername())
	if username == "" {
		return fmt.Errorf("empty target user")
	}

	currentShell := strings.TrimSpace(getDefaultShell(username))
	if strings.HasSuffix(currentShell, "/zsh") {
		return nil
	}

	logx.Info("setting default shell", "user", username, "shell", zshPath)
	// Try chsh first
	if err := runCommand("chsh", "-s", zshPath, username); err != nil {
		logx.Warn("chsh failed; falling back to usermod", "err", err)
		// Fallback to usermod
		if err2 := runCommand("usermod", "-s", zshPath, username); err2 != nil {
			logx.Error("usermod failed setting shell", "err", err2)
			return fmt.Errorf("usermod set shell: %w", err2)
		}
	}

	// Verify (allow a second check in case of delayed passwd update)
	for i := 0; i < 2; i++ {
		newShell := strings.TrimSpace(getDefaultShell(username))
		if strings.HasSuffix(newShell, "/zsh") {
			logx.Info("default shell set to zsh", "user", username)
			return nil
		}
	}
	logx.Error("verification failed setting shell", "user", username, "expected", "zsh", "got", strings.TrimSpace(getDefaultShell(username)))
	return fmt.Errorf("verification failed: expected zsh, got %q", strings.TrimSpace(getDefaultShell(username)))
}
