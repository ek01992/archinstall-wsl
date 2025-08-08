package dotfiles

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// NOTE: Package-level seams below are for testability and are NOT concurrency-safe.
// Use internal/seams.With in tests to serialize overrides. Prefer DI if adding concurrency.
var (
	getUserHomeDir = func() (string, error) { return osUserHomeDir() }
	osUserHomeDir  = func() (string, error) { return osUserHomeDirImpl() }

	// File/FS seams
	pathExists = func(path string) bool {
		_, err := os.Stat(path)
		return err == nil
	}
	writeFile = func(path string, data []byte, perm fs.FileMode) error { return os.WriteFile(path, data, perm) }
	listFiles = func(dir string) ([]string, error) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return nil, err
		}
		names := make([]string, 0, len(entries))
		for _, e := range entries {
			names = append(names, e.Name())
		}
		return names, nil
	}
	lstat    = func(path string) (fs.FileMode, error) { fi, err := os.Lstat(path); if err != nil { return 0, err }; return fi.Mode(), nil }
	readlink = func(path string) (string, error) { return os.Readlink(path) }
	symlink  = func(oldname string, newname string) error { return os.Symlink(oldname, newname) }

	// Command seam
	runCommand = func(name string, args ...string) error { return exec.Command(name, args...).Run() }
)

// osUserHomeDirImpl is split so tests can override getUserHomeDir only.
func osUserHomeDirImpl() (string, error) { return os.UserHomeDir() }

// installDotfiles clones repo into ~/.dotfiles when repoURL is provided, then symlinks files
// to the home directory (skipping README and dot git directories). When repoURL is empty,
// it writes a default .zshrc.
func installDotfiles(repoURL string) error {
	home, err := getUserHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		return fmt.Errorf("cannot determine home directory: %w", err)
	}

	repoURL = strings.TrimSpace(repoURL)
	if repoURL == "" {
		// Write default zshrc
		defaultContent := `export ZSH="$HOME/.oh-my-zsh"
ZSH_THEME="robbyrussell"
plugins=(git)
source $ZSH/oh-my-zsh.sh
`
		z := filepath.Join(home, ".zshrc")
		if err := writeFile(z, []byte(defaultContent), 0o644); err != nil {
			return fmt.Errorf("write default .zshrc: %w", err)
		}
		return nil
	}

	repoDir := filepath.Join(home, ".dotfiles")
	if !pathExists(repoDir) {
		if err := runCommand("git", "clone", "--depth", "1", repoURL, repoDir); err != nil {
			return fmt.Errorf("git clone: %w", err)
		}
	}

	items, err := listFiles(repoDir)
	if err != nil {
		return fmt.Errorf("list repo files: %w", err)
	}

	for _, name := range items {
		base := filepath.Base(name)
		if base == "README.md" || base == ".git" {
			continue
		}

		src := filepath.Join(repoDir, name)
		destName := base
		// if file name already starts with dot, keep; else map zshrc -> .zshrc
		if !strings.HasPrefix(destName, ".") {
			switch destName {
			case "zshrc":
				destName = ".zshrc"
			default:
				// keep non-hidden names as-is but in $HOME
			}
		}
		dest := filepath.Join(home, destName)

		mode, err := lstat(dest)
		if err == nil && mode&fs.ModeSymlink == fs.ModeSymlink {
			link, err := readlink(dest)
			if err == nil && link == src {
				// Already correct
				continue
			}
		}

		if err := symlink(src, dest); err != nil {
			return fmt.Errorf("symlink %s -> %s: %w", dest, src, err)
		}
	}

	return nil
}
