package user

import (
	"fmt"
	"os"
	stduser "os/user"
	"path/filepath"
	"strings"
)

var (
	getHomeDirByUsername = func(username string) (string, error) {
		u, err := stduser.Lookup(username)
		if err != nil {
			return "", err
		}
		return u.HomeDir, nil
	}
	pathExists = func(path string) bool {
		_, err := os.Stat(path)
		return err == nil
	}
)

func buildZshrc(theme string, plugins []string) string {
	pluginLine := strings.Join(plugins, " ")
	// Keep minimal content required by tests in a stable order
	b := &strings.Builder{}
	b.WriteString("export ZSH=\"$HOME/.oh-my-zsh\"\n")
	b.WriteString(fmt.Sprintf("ZSH_THEME=\"%s\"\n", strings.TrimSpace(theme)))
	b.WriteString(fmt.Sprintf("plugins=(%s)\n", strings.TrimSpace(pluginLine)))
	b.WriteString("source $ZSH/oh-my-zsh.sh\n")
	return b.String()
}

// installOhMyZsh installs oh-my-zsh for the given user and ensures .zshrc reflects
// the provided theme and plugins. Idempotent: no writes if already in desired state.
func installOhMyZsh(username string, theme string, plugins []string) error {
	username = strings.TrimSpace(username)
	if username == "" {
		return fmt.Errorf("username must not be empty")
	}

	home, err := getHomeDirByUsername(username)
	if err != nil || strings.TrimSpace(home) == "" {
		return fmt.Errorf("failed to resolve home for %s: %v", username, err)
	}

	omzDir := filepath.Join(home, ".oh-my-zsh")
	if !pathExists(omzDir) {
		// Clone oh-my-zsh if missing
		if err := runCommand("git", "clone", "--depth", "1", "https://github.com/ohmyzsh/ohmyzsh.git", omzDir); err != nil {
			return fmt.Errorf("clone oh-my-zsh: %w", err)
		}
	}

	desired := buildZshrc(theme, plugins)
	zshrc := filepath.Join(home, ".zshrc")

	current, err := readFile(zshrc)
	if err == nil {
		if string(current) == desired {
			return nil
		}
	}

	if err := writeFile(zshrc, []byte(desired), 0o644); err != nil {
		return fmt.Errorf("write .zshrc: %w", err)
	}

	// Verify
	verify, err := readFile(zshrc)
	if err != nil || string(verify) != desired {
		return fmt.Errorf("verification failed: .zshrc not in desired state")
	}

	return nil
}
