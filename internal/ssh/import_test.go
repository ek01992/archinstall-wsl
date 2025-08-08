package ssh

import (
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
)

func TestImportSSHKeys_CreatesDirSetsPermsAndCopiesKeys(t *testing.T) {
	origHome := getUserHomeDir
	origList := listFiles
	origRead := readFile
	origWrite := writeFile
	origMkdir := mkdirAll
	origChmod := chmod

	t.Cleanup(func() {
		getUserHomeDir = origHome
		listFiles = origList
		readFile = origRead
		writeFile = origWrite
		mkdirAll = origMkdir
		chmod = origChmod
	})

	getUserHomeDir = func() (string, error) { return "/home/alice", nil }

	host := "/mnt/c/Users/Alice/.ssh"
	listFiles = func(dir string) ([]string, error) {
		if dir != host {
			t.Fatalf("unexpected listFiles dir: %q", dir)
		}
		return []string{"id_ed25519", "id_ed25519.pub"}, nil
	}

	readFile = func(path string) ([]byte, error) {
		if strings.HasPrefix(path, host) {
			base := filepath.Base(path)
			switch base {
			case "id_ed25519":
				return []byte("PRIVATE-KEY"), nil
			case "id_ed25519.pub":
				return []byte("ssh-ed25519 AAA... alice@pc\n"), nil
			}
		}
		// Destination does not exist yet
		return nil, fs.ErrNotExist
	}

	var madeDir string
	var madeMode fs.FileMode
	mkdirAll = func(path string, perm fs.FileMode) error { madeDir = path; madeMode = perm; return nil }

	chmodCalls := map[string]fs.FileMode{}
	chmod = func(path string, mode fs.FileMode) error { chmodCalls[path] = mode; return nil }

	writes := map[string]struct {
		data string
		mode fs.FileMode
	}{}
	writeFile = func(path string, data []byte, perm fs.FileMode) error {
		writes[path] = struct {
			data string
			mode fs.FileMode
		}{string(data), perm}
		return nil
	}

	if err := importSSHKeysFromWindows(host); err != nil {
		t.Fatalf("importSSHKeysFromWindows returned error: %v", err)
	}

	if madeDir != "/home/alice/.ssh" || madeMode != 0o700 {
		t.Fatalf("expected mkdirAll with 0700 on ~/.ssh; got path=%q mode=%v", madeDir, madeMode)
	}
	if chmodCalls["/home/alice/.ssh"] != 0o700 {
		t.Fatalf("expected chmod 0700 on ~/.ssh")
	}

	priv := "/home/alice/.ssh/id_ed25519"
	pub := "/home/alice/.ssh/id_ed25519.pub"

	if w, ok := writes[priv]; !ok || w.data != "PRIVATE-KEY" || w.mode != 0o600 {
		t.Fatalf("expected write of private key 0600; got %+v", w)
	}
	if w, ok := writes[pub]; !ok || !strings.HasPrefix(w.data, "ssh-ed25519 ") || w.mode != 0o644 {
		t.Fatalf("expected write of public key 0644; got %+v", w)
	}

	if chmodCalls[priv] != 0o600 {
		t.Fatalf("expected chmod 0600 on private key")
	}
	if chmodCalls[pub] != 0o644 {
		t.Fatalf("expected chmod 0644 on public key")
	}
}

func TestImportSSHKeys_Idempotent_NoRewriteWhenIdentical(t *testing.T) {
	origHome := getUserHomeDir
	origList := listFiles
	origRead := readFile
	origWrite := writeFile
	origMkdir := mkdirAll
	origChmod := chmod

	t.Cleanup(func() {
		getUserHomeDir = origHome
		listFiles = origList
		readFile = origRead
		writeFile = origWrite
		mkdirAll = origMkdir
		chmod = origChmod
	})

	getUserHomeDir = func() (string, error) { return "/home/bob", nil }
	host := "/mnt/c/Users/Bob/.ssh"

	listFiles = func(dir string) ([]string, error) { return []string{"id_rsa"}, nil }

	readFile = func(path string) ([]byte, error) {
		base := filepath.Base(path)
		switch base {
		case "id_rsa":
			return []byte("SAME"), nil
		}
		return nil, fs.ErrNotExist
	}

	mkdirAll = func(path string, perm fs.FileMode) error { return nil }

	writeCalled := false
	writeFile = func(path string, data []byte, perm fs.FileMode) error { writeCalled = true; return nil }

	chmod = func(path string, mode fs.FileMode) error { return nil }

	if err := importSSHKeysFromWindows(host); err != nil {
		t.Fatalf("importSSHKeysFromWindows returned error: %v", err)
	}
	if writeCalled {
		t.Fatalf("did not expect write when contents are identical (idempotent)")
	}
}

func TestImportSSHKeys_EmptyHostPathError_NoCalls(t *testing.T) {
	origHome := getUserHomeDir
	origList := listFiles
	origRead := readFile
	origWrite := writeFile
	origMkdir := mkdirAll
	origChmod := chmod

	t.Cleanup(func() {
		getUserHomeDir = origHome
		listFiles = origList
		readFile = origRead
		writeFile = origWrite
		mkdirAll = origMkdir
		chmod = origChmod
	})

	getUserHomeDir = func() (string, error) { t.Fatalf("home should not be called"); return "", nil }
	listFiles = func(dir string) ([]string, error) { t.Fatalf("listFiles should not be called"); return nil, nil }
	readFile = func(path string) ([]byte, error) { t.Fatalf("readFile should not be called"); return nil, nil }
	writeFile = func(path string, data []byte, perm fs.FileMode) error {
		t.Fatalf("writeFile should not be called")
		return nil
	}
	mkdirAll = func(path string, perm fs.FileMode) error { t.Fatalf("mkdirAll should not be called"); return nil }
	chmod = func(path string, mode fs.FileMode) error { t.Fatalf("chmod should not be called"); return nil }

	if err := importSSHKeysFromWindows("   \t\n"); err == nil {
		t.Fatalf("expected error for empty host path")
	}
}
