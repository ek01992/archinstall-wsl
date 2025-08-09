package user

import (
	"fmt"
	"strings"

	"archwsl-tui-configurator/internal/tx"
)

// getDefaultShellFromPasswd parses /etc/passwd content and returns the shell for the given username.
func getDefaultShellFromPasswd(passwdContent []byte, username string) string {
	username = strings.TrimSpace(username)
	if username == "" {
		return ""
	}
	for _, line := range strings.Split(string(passwdContent), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, ":")
		if len(parts) < 7 {
			continue
		}
		if parts[0] != username {
			continue
		}
		return strings.TrimSpace(parts[6])
	}
	return ""
}

func (s *Service) currentShell(username string) string {
	if b, err := s.fs.ReadFile("/etc/passwd"); err == nil {
		return strings.TrimSpace(getDefaultShellFromPasswd(b, username))
	}
	return ""
}

// InstallZsh sets the current user's login shell to zsh, idempotently.
func (s *Service) InstallZsh() error {
	username := strings.TrimSpace(s.id.CurrentUsername())
	if username == "" {
		return fmt.Errorf("empty target user")
	}
	// Idempotency via fs
	if cs := s.currentShell(username); strings.HasSuffix(cs, "/zsh") {
		return nil
	}
	// Try chsh first, then fallback to usermod
	if err := s.cmd.Run("chsh", "-s", s.zshPath, username); err != nil {
		if err2 := s.cmd.Run("usermod", "-s", s.zshPath, username); err2 != nil {
			return fmt.Errorf("usermod set shell: %w", err2)
		}
	}
	// Verify using fs with a brief retry
	for i := 0; i < 2; i++ {
		if ns := s.currentShell(username); strings.HasSuffix(ns, "/zsh") {
			return nil
		}
	}
	return fmt.Errorf("verification failed: expected zsh, got %q", s.currentShell(username))
}

// InstallZshTx sets zsh and restores previous shell on failure.
func (s *Service) InstallZshTx() (err error) {
	tr := tx.New()
	defer func() {
		if err != nil {
			_ = tr.Rollback()
		}
	}()

	username := strings.TrimSpace(s.id.CurrentUsername())
	if username == "" {
		return fmt.Errorf("empty target user")
	}
	prevShell := s.currentShell(username)
	if prevShell != "" {
		u := username
		p := prevShell
		tr.Defer(func() error { return s.cmd.Run("chsh", "-s", p, u) })
	}
	if err = s.InstallZsh(); err != nil {
		return err
	}
	tr.Commit()
	return nil
}
