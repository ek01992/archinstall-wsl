package ssh

import (
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
)

func TestImportSSHKeysFromWindowsTx_RollsBackOnFailure(t *testing.T) {
	origHome := getUserHomeDir
	origList := listFiles
	origRead := readFile
	origWrite := writeFile
	origMkdir := mkdirAll
	origChmod := chmod
	origRemove := removeFile

	t.Cleanup(func() {
		getUserHomeDir = origHome
		listFiles = origList
		readFile = origRead
		writeFile = origWrite
		mkdirAll = origMkdir
		chmod = origChmod
		removeFile = origRemove
	})

	getUserHomeDir = func() (string, error) { return "/home/alice", nil }
	listFiles = func(dir string) ([]string, error) { return []string{"id_ed25519", "id_ed25519.pub"}, nil }

	readFile = func(path string) ([]byte, error) {
		if strings.Contains(path, "/home/alice/.ssh/") {
			return nil, fs.ErrNotExist
		}
		return []byte("SRC"), nil
	}
	removed := map[string]bool{}
	removeFile = func(path string) error { removed[path] = true; return nil }

	mkdirAll = func(path string, perm fs.FileMode) error { return nil }
	chmod = func(path string, mode fs.FileMode) error { return nil }
	writeFile = func(path string, data []byte, perm fs.FileMode) error {
		// Induce failure on first write to trigger rollback
		return fs.ErrPermission
	}

	if err := importSSHKeysFromWindowsTx("/host/.ssh"); err == nil {
		t.Fatalf("expected failure to trigger rollback")
	}
	if !removed[filepath.Join("/home/alice/.ssh", "id_ed25519")] || !removed[filepath.Join("/home/alice/.ssh", "id_ed25519.pub")] {
		t.Fatalf("expected rollback to remove newly created files")
	}
}

func TestImportSSHKeysFromWindowsTx_SuccessDoesNotRollback(t *testing.T) {
	origHome := getUserHomeDir
	origList := listFiles
	origRead := readFile
	origWrite := writeFile
	origMkdir := mkdirAll
	origChmod := chmod
	origRemove := removeFile

	t.Cleanup(func() {
		getUserHomeDir = origHome
		listFiles = origList
		readFile = origRead
		writeFile = origWrite
		mkdirAll = origMkdir
		chmod = origChmod
		removeFile = origRemove
	})

	getUserHomeDir = func() (string, error) { return "/home/bob", nil }
	listFiles = func(dir string) ([]string, error) { return []string{"id_rsa"}, nil }
	readFile = func(path string) ([]byte, error) {
		if strings.Contains(path, "/home/bob/.ssh/") {
			return nil, fs.ErrNotExist
		}
		return []byte("SRC"), nil
	}
	mkdirAll = func(path string, perm fs.FileMode) error { return nil }
	chmod = func(path string, mode fs.FileMode) error { return nil }
	writeFile = func(path string, data []byte, perm fs.FileMode) error { return nil }

	removed := false
	removeFile = func(path string) error { removed = true; return nil }

	if err := importSSHKeysFromWindowsTx("/host/.ssh"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if removed {
		t.Fatalf("did not expect rollback on success")
	}
}
