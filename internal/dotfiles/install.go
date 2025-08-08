package dotfiles

import (
    "fmt"
    "io/fs"
    "path/filepath"
    "strings"
)

var (
    getUserHomeDir = func() (string, error) { return osUserHomeDir() }
    osUserHomeDir = func() (string, error) { return osUserHomeDirImpl() }

    // File/FS seams
    pathExists = func(path string) bool {
        _, err := readFile(path)
        return err == nil
    }
    readFile  = func(path string) ([]byte, error) { return nil, fs.ErrNotExist }
    writeFile = func(path string, data []byte, perm fs.FileMode) error { return nil }
    listFiles = func(dir string) ([]string, error) { return nil, nil }
    lstat     = func(path string) (fs.FileMode, error) { return 0, fs.ErrNotExist }
    readlink  = func(path string) (string, error) { return "", fs.ErrNotExist }
    symlink   = func(oldname, newname string) error { return nil }

    // Command seam
    runCommand = func(name string, args ...string) error { return nil }
)

// osUserHomeDirImpl is split so tests can override getUserHomeDir only.
func osUserHomeDirImpl() (string, error) { return "", fmt.Errorf("not implemented") }

// installDotfiles clones repo into ~/.dotfiles when repoURL is provided, then symlinks files
// to the home directory (skipping README and dot git directories). When repoURL is empty,
// it writes a default .zshrc.
func installDotfiles(repoURL string) error {
    home, err := getUserHomeDir()
    if err != nil || strings.TrimSpace(home) == "" {
        return fmt.Errorf("cannot determine home directory: %v", err)
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
