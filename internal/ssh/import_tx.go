package ssh

import (
	"io/fs"
	"path/filepath"

	"archwsl-tui-configurator/internal/tx"
)

// importSSHKeysFromWindowsTx wraps import with rollback for each file.
func importSSHKeysFromWindowsTx(hostPath string) (err error) {
	tr := tx.New()
	defer func() {
		if err != nil {
			_ = tr.Rollback()
		}
	}()

	home, herr := getUserHomeDir()
	if herr != nil {
		return herr
	}
	dotSSH := filepath.Join(home, ".ssh")
	// For each file, capture previous state and register undo
	names, lerr := listFiles(hostPath)
	if lerr != nil {
		return lerr
	}
	for _, name := range names {
		dst := filepath.Join(dotSSH, name)
		prev, perr := readFile(dst)
		if perr == nil {
			path := dst
			data := append([]byte(nil), prev...)
			// Attempt to preserve previous mode; fall back to 0600
			mode := fs.FileMode(0o600)
			if m, statErr := lstatMode(path); statErr == nil {
				mode = m
			}
			m := mode
			tr.Defer(func() error {
				if err := writeFile(path, data, m); err != nil { return err }
				return chmod(path, m)
			})
		} else {
			path := dst
			tr.Defer(func() error { return removeFile(path) })
		}
	}

	if err = importSSHKeysFromWindows(hostPath); err != nil {
		return err
	}
	tr.Commit()
	return nil
}

var removeFile = func(path string) error { return nil }

// lstatMode is a seam to obtain a file's permission bits
var lstatMode = func(path string) (fs.FileMode, error) { return fs.FileMode(0), fs.ErrNotExist }
