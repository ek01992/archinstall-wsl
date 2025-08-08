package user

import (
	"fmt"
	"path/filepath"
	"strings"
)

// InstallOhMyZsh installs oh-my-zsh and writes ~/.zshrc for the target username.
func (s *Service) InstallOhMyZsh(username string, theme string, plugins []string) error {
	username = strings.TrimSpace(username)
	if username == "" {
		return fmt.Errorf("username must not be empty")
	}
	home, err := s.id.HomeDirByUsername(username)
	if err != nil || strings.TrimSpace(home) == "" {
		return fmt.Errorf("failed to resolve home for %s: %w", username, err)
	}
	omzDir := filepath.Join(home, ".oh-my-zsh")
	// Best-effort existence check: attempt to read dir listing; if err, assume missing and clone
	if _, err := s.fs.ReadFile(filepath.Join(omzDir, ".keep")); err != nil {
		if err := s.cmd.Run("git", "clone", "--depth", "1", "https://github.com/ohmyzsh/ohmyzsh.git", omzDir); err != nil {
			return fmt.Errorf("clone oh-my-zsh: %w", err)
		}
	}
	desired := buildZshrc(theme, plugins)
	zshrc := filepath.Join(home, ".zshrc")
	if cur, err := s.fs.ReadFile(zshrc); err == nil && string(cur) == desired {
		return nil
	}
	if err := s.fs.WriteFile(zshrc, []byte(desired), 0o644); err != nil {
		return fmt.Errorf("write .zshrc: %w", err)
	}
	if verify, err := s.fs.ReadFile(zshrc); err != nil || string(verify) != desired {
		return fmt.Errorf("verification failed: .zshrc not in desired state")
	}
	return nil
}
