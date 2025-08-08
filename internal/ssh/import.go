package ssh

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var (
	getUserHomeDir = func() (string, error) { return os.UserHomeDir() }
	listFiles      = func(dir string) ([]string, error) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return nil, err
		}
		names := make([]string, 0, len(entries))
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			names = append(names, e.Name())
		}
		return names, nil
	}
	readFile  = func(path string) ([]byte, error) { return os.ReadFile(path) }
	writeFile = func(path string, data []byte, perm fs.FileMode) error { return os.WriteFile(path, data, perm) }
	mkdirAll  = func(path string, perm fs.FileMode) error { return os.MkdirAll(path, perm) }
	chmod     = func(path string, mode fs.FileMode) error { return os.Chmod(path, mode) }
)

// importSSHKeysFromWindows copies SSH keys from a Windows host path into the user's
// ~/.ssh directory, applying correct permissions. The operation is idempotent:
// existing files with identical content are not rewritten.
func importSSHKeysFromWindows(hostPath string) error {
	hostPath = strings.TrimSpace(hostPath)
	if hostPath == "" {
		return errors.New("hostPath must not be empty")
	}

	home, err := getUserHomeDir()
	if err != nil {
		return fmt.Errorf("resolve home dir: %w", err)
	}
	if strings.TrimSpace(home) == "" {
		return errors.New("empty home directory")
	}

	dotSSH := filepath.Join(home, ".ssh")
	if err := mkdirAll(dotSSH, 0o700); err != nil {
		return fmt.Errorf("ensure ~/.ssh: %w", err)
	}
	// Best-effort to enforce directory permission
	_ = chmod(dotSSH, 0o700)

	names, err := listFiles(hostPath)
	if err != nil {
		return fmt.Errorf("list host ssh dir: %w", err)
	}

	for _, name := range names {
		if name == "" || name == "." || name == ".." {
			continue
		}
		src := filepath.Join(hostPath, name)
		dst := filepath.Join(dotSSH, name)

		srcBytes, err := readFile(src)
		if err != nil {
			return fmt.Errorf("read source %s: %w", src, err)
		}

		mode := fs.FileMode(0o600)
		if strings.HasSuffix(name, ".pub") {
			mode = 0o644
		}

		dstBytes, err := readFile(dst)
		if err == nil {
			if bytes.Equal(dstBytes, srcBytes) {
				// Already in desired state; ensure perms and continue
				_ = chmod(dst, mode)
				continue
			}
		}

		if err := writeFile(dst, srcBytes, mode); err != nil {
			return fmt.Errorf("write destination %s: %w", dst, err)
		}
		_ = chmod(dst, mode)
	}

	return nil
}
