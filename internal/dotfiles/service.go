package dotfiles

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

type Runner interface { Run(name string, args ...string) error }

type FS interface {
	UserHomeDir() (string, error)
	WriteFile(path string, data []byte, perm fs.FileMode) error
	ReadDir(path string) ([]fs.DirEntry, error)
	Lstat(path string) (fs.FileInfo, error)
	Readlink(path string) (string, error)
	Symlink(oldname, newname string) error
}

type Service struct { fs FS; r Runner }

func NewService(fs FS, r Runner) *Service { return &Service{fs: fs, r: r} }

// Install clones repo into ~/.dotfiles when repoURL is provided, then symlinks files
// to the home directory (skipping README and .git). When repoURL is empty, it writes a default .zshrc.
func (s *Service) Install(repoURL string) error {
	home, err := s.fs.UserHomeDir()
	if err != nil || strings.TrimSpace(home) == "" { return fmt.Errorf("cannot determine home directory: %w", err) }
	repoURL = strings.TrimSpace(repoURL)
	if repoURL == "" {
		defaultContent := `export ZSH="$HOME/.oh-my-zsh"
ZSH_THEME="robbyrussell"
plugins=(git)
source $ZSH/oh-my-zsh.sh
`
		z := filepath.Join(home, ".zshrc")
		if err := s.fs.WriteFile(z, []byte(defaultContent), 0o644); err != nil { return fmt.Errorf("write default .zshrc: %w", err) }
		return nil
	}
	repoDir := filepath.Join(home, ".dotfiles")
	if _, err := s.fs.Lstat(repoDir); err != nil {
		if err := s.r.Run("git", "clone", "--depth", "1", repoURL, repoDir); err != nil { return fmt.Errorf("git clone: %w", err) }
	}
	entries, err := s.fs.ReadDir(repoDir); if err != nil { return fmt.Errorf("list repo files: %w", err) }
	for _, e := range entries {
		name := e.Name()
		base := filepath.Base(name)
		if base == "README.md" || base == ".git" { continue }
		src := filepath.Join(repoDir, name)
		destName := base
		if !strings.HasPrefix(destName, ".") {
			switch destName { case "zshrc": destName = ".zshrc" }
		}
		dest := filepath.Join(home, destName)
		if fi, err := s.fs.Lstat(dest); err == nil && (fi.Mode()&fs.ModeSymlink) == fs.ModeSymlink {
			if link, err := s.fs.Readlink(dest); err == nil && link == src { continue }
		}
		if err := s.fs.Symlink(src, dest); err != nil { return fmt.Errorf("symlink %s -> %s: %w", dest, src, err) }
	}
	return nil
}
