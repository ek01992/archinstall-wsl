package user

import (
	"errors"
	"io/fs"
	"strings"
	"testing"
)

func TestCreateUser_CreatesAddsWheelSetsPasswordAndSudoers(t *testing.T) {
	// Save and restore seams
	origLookup := lookupUserByName
	origRun := runCommand
	origRunWithStdin := runCommandWithStdin
	origRead := readFile
	origWrite := writeFile
	origSudoersD := sudoersDPath

	t.Cleanup(func() {
		lookupUserByName = origLookup
		runCommand = origRun
		runCommandWithStdin = origRunWithStdin
		readFile = origRead
		writeFile = origWrite
		sudoersDPath = origSudoersD
	})

	// User does not exist initially
	lookupUserByName = func(name string) (any, error) {
		if name != "alice" {
			t.Fatalf("expected lookup for 'alice', got %q", name)
		}
		return nil, errors.New("not found")
	}

	var useraddCalled, usermodCalled, chpasswdCalled bool

	runCommand = func(name string, args ...string) error {
		switch name {
		case "useradd":
			if len(args) != 2 || args[0] != "-m" || args[1] != "alice" {
				t.Fatalf("unexpected useradd args: %v", args)
			}
			useraddCalled = true
			return nil
		case "usermod":
			if len(args) != 3 || args[0] != "-aG" || args[1] != "wheel" || args[2] != "alice" {
				t.Fatalf("unexpected usermod args: %v", args)
			}
			usermodCalled = true
			return nil
		default:
			t.Fatalf("unexpected command %q with args %v", name, args)
			return nil
		}
	}

	runCommandWithStdin = func(name string, stdin string, args ...string) error {
		if name != "chpasswd" {
			t.Fatalf("expected chpasswd, got %q", name)
		}
		if len(args) != 0 {
			t.Fatalf("expected no extra args for chpasswd, got %v", args)
		}
		if stdin != "alice:secret" {
			t.Fatalf("unexpected chpasswd stdin: %q", stdin)
		}
		chpasswdCalled = true
		return nil
	}

	sudoersDPath = "/fake/etc/sudoers.d"

	readFile = func(path string) ([]byte, error) {
		if !strings.HasSuffix(path, "/010_wheel_nopasswd") {
			t.Fatalf("unexpected sudoers file read path: %q", path)
		}
		return nil, fs.ErrNotExist
	}

	var wrotePath string
	var wrotePerm fs.FileMode
	var wroteContent string

	writeFile = func(path string, data []byte, perm fs.FileMode) error {
		wrotePath = path
		wrotePerm = perm
		wroteContent = string(data)
		return nil
	}

	if err := createUser("alice", "secret"); err != nil {
		t.Fatalf("createUser returned error: %v", err)
	}

	if !useraddCalled {
		t.Fatalf("expected useradd to be called")
	}
	if !usermodCalled {
		t.Fatalf("expected usermod to be called")
	}
	if !chpasswdCalled {
		t.Fatalf("expected chpasswd to be called")
	}

	if wrotePath != "/fake/etc/sudoers.d/010_wheel_nopasswd" {
		t.Fatalf("unexpected sudoers write path: %q", wrotePath)
	}
	if wrotePerm != 0o440 {
		t.Fatalf("unexpected sudoers file mode: got %v want 0440", wrotePerm)
	}
	if strings.TrimSpace(wroteContent) != "%wheel ALL=(ALL) NOPASSWD: ALL" {
		t.Fatalf("unexpected sudoers content: %q", wroteContent)
	}
}

func TestCreateUser_IdempotentWhenUserExists_NoUserAdd_NoSudoersRewrite(t *testing.T) {
	// Save and restore seams
	origLookup := lookupUserByName
	origRun := runCommand
	origRunWithStdin := runCommandWithStdin
	origRead := readFile
	origWrite := writeFile
	origSudoersD := sudoersDPath

	t.Cleanup(func() {
		lookupUserByName = origLookup
		runCommand = origRun
		runCommandWithStdin = origRunWithStdin
		readFile = origRead
		writeFile = origWrite
		sudoersDPath = origSudoersD
	})

	// User already exists
	lookupUserByName = func(name string) (any, error) { return struct{}{}, nil }

	var useraddCalled, usermodCalled, chpasswdCalled, wroteCalled bool

	runCommand = func(name string, args ...string) error {
		switch name {
		case "useradd":
			useraddCalled = true
			return errors.New("should not be called")
		case "usermod":
			if len(args) == 3 && args[0] == "-aG" && args[1] == "wheel" && args[2] == "bob" {
				usermodCalled = true
				return nil
			}
			t.Fatalf("unexpected usermod args: %v", args)
			return nil
		default:
			t.Fatalf("unexpected command %q", name)
			return nil
		}
	}

	runCommandWithStdin = func(name string, stdin string, args ...string) error {
		if name != "chpasswd" {
			t.Fatalf("expected chpasswd, got %q", name)
		}
		if stdin != "bob:s3cr3t" {
			t.Fatalf("unexpected chpasswd stdin: %q", stdin)
		}
		chpasswdCalled = true
		return nil
	}

	sudoersDPath = "/fake/etc/sudoers.d"

	readFile = func(path string) ([]byte, error) {
		if !strings.HasSuffix(path, "/010_wheel_nopasswd") {
			t.Fatalf("unexpected sudoers file read path: %q", path)
		}
		return []byte("%wheel ALL=(ALL) NOPASSWD: ALL\n"), nil
	}

	writeFile = func(path string, data []byte, perm fs.FileMode) error {
		wroteCalled = true
		return errors.New("should not write when content matches")
	}

	if err := createUser("bob", "s3cr3t"); err != nil {
		t.Fatalf("createUser returned error: %v", err)
	}

	if useraddCalled {
		t.Fatalf("did not expect useradd when user exists")
	}
	if !usermodCalled {
		t.Fatalf("expected usermod to be called to ensure wheel membership")
	}
	if !chpasswdCalled {
		t.Fatalf("expected chpasswd to be called to set password")
	}
	if wroteCalled {
		t.Fatalf("did not expect sudoers file to be rewritten when identical")
	}
}

func TestCreateUser_EmptyUsernameReturnsErrorAndNoActions(t *testing.T) {
	// Save and restore seams
	origLookup := lookupUserByName
	origRun := runCommand
	origRunWithStdin := runCommandWithStdin
	origRead := readFile
	origWrite := writeFile

	t.Cleanup(func() {
		lookupUserByName = origLookup
		runCommand = origRun
		runCommandWithStdin = origRunWithStdin
		readFile = origRead
		writeFile = origWrite
	})

	lookupUserByName = func(name string) (any, error) {
		t.Fatalf("lookup should not be called for empty username")
		return nil, nil
	}
	runCommand = func(name string, args ...string) error {
		t.Fatalf("no commands should be run for empty username")
		return nil
	}
	runCommandWithStdin = func(name string, stdin string, args ...string) error {
		t.Fatalf("no commands should be run for empty username")
		return nil
	}
	readFile = func(path string) ([]byte, error) {
		t.Fatalf("no files should be read for empty username")
		return nil, nil
	}
	writeFile = func(path string, data []byte, perm fs.FileMode) error {
		t.Fatalf("no files should be written for empty username")
		return nil
	}

	if err := createUser("  \t\n", "whatever"); err == nil {
		t.Fatalf("expected error for empty/whitespace username")
	}
}

func TestCreateUser_FallbackToGpasswdWhenUsermodFails(t *testing.T) {
	// Save and restore seams
	origLookup := lookupUserByName
	origRun := runCommand
	origRunWithStdin := runCommandWithStdin
	origRead := readFile
	origWrite := writeFile
	origSudoersD := sudoersDPath

	t.Cleanup(func() {
		lookupUserByName = origLookup
		runCommand = origRun
		runCommandWithStdin = origRunWithStdin
		readFile = origRead
		writeFile = origWrite
		sudoersDPath = origSudoersD
	})

	// User exists to skip useradd
	lookupUserByName = func(name string) (any, error) { return struct{}{}, nil }

	var usermodCalled, gpasswdCalled bool

	runCommand = func(name string, args ...string) error {
		switch name {
		case "usermod":
			usermodCalled = true
			return errors.New("usermod failed")
		case "gpasswd":
			if len(args) != 3 || args[0] != "-a" || args[2] != "wheel" {
				t.Fatalf("unexpected gpasswd args: %v", args)
			}
			gpasswdCalled = true
			return nil
		default:
			t.Fatalf("unexpected command %q", name)
			return nil
		}
	}

	runCommandWithStdin = func(name string, stdin string, args ...string) error { return nil }

	sudoersDPath = "/fake/etc/sudoers.d"
	readFile = func(path string) ([]byte, error) { return nil, fs.ErrNotExist }
	writeFile = func(path string, data []byte, perm fs.FileMode) error { return nil }

	if err := createUser("carol", ""); err != nil {
		t.Fatalf("createUser returned error: %v", err)
	}

	if !usermodCalled {
		t.Fatalf("expected usermod to be attempted")
	}
	if !gpasswdCalled {
		t.Fatalf("expected gpasswd fallback to be called when usermod fails")
	}
}

func TestCreateUser_CreatesSudoersDirIfMissing(t *testing.T) {
	origLookup := lookupUserByName
	origRun := runCommand
	origRunWithStdin := runCommandWithStdin
	origRead := readFile
	origWrite := writeFile
	origSudoersD := sudoersDPath
	origMkdir := mkdirAll

	t.Cleanup(func() {
		lookupUserByName = origLookup
		runCommand = origRun
		runCommandWithStdin = origRunWithStdin
		readFile = origRead
		writeFile = origWrite
		sudoersDPath = origSudoersD
		mkdirAll = origMkdir
	})

	lookupUserByName = func(name string) (any, error) { return struct{}{}, nil }

	sudoersDPath = "/fake/etc/sudoers.d"
	mkdirCalled := false
	mkdirPath := ""
	mkdirPerm := fs.FileMode(0)
	mkdirAll = func(path string, perm fs.FileMode) error { mkdirCalled = true; mkdirPath = path; mkdirPerm = perm; return nil }

	readFile = func(path string) ([]byte, error) { return nil, fs.ErrNotExist }
	writeFile = func(path string, data []byte, perm fs.FileMode) error { return nil }
	runCommand = func(name string, args ...string) error { return nil }
	runCommandWithStdin = func(name string, stdin string, args ...string) error { return nil }

	if err := createUser("dave", "pw"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !mkdirCalled || mkdirPath != "/fake/etc/sudoers.d" || mkdirPerm != 0o755 {
		t.Fatalf("expected sudoers dir mkdirAll 0755; got called=%v path=%q perm=%v", mkdirCalled, mkdirPath, mkdirPerm)
	}
}

func TestCreateUser_InvalidCredentialsReturnsError(t *testing.T) {
	origRun := runCommand
	origRunStdin := runCommandWithStdin
	t.Cleanup(func() { runCommand = origRun; runCommandWithStdin = origRunStdin })

	runCommand = func(name string, args ...string) error { t.Fatalf("should not run commands"); return nil }
	runCommandWithStdin = func(name string, stdin string, args ...string) error { t.Fatalf("should not set password"); return nil }

	cases := [][2]string{{"al:ice", "pw"}, {"alice", "bad\nline"}}
	for _, c := range cases {
		if err := createUser(c[0], c[1]); err == nil {
			t.Fatalf("expected error for invalid inputs: %v", c)
		}
	}
}

func TestCreateUser_SudoersValidationFailurePreventsWrite(t *testing.T) {
	origLookup := lookupUserByName
	origRun := runCommand
	origRunWithStdin := runCommandWithStdin
	origRead := readFile
	origWrite := writeFile
	origValidate := validateSudoersContent
	origSudoersD := sudoersDPath

	t.Cleanup(func() {
		lookupUserByName = origLookup
		runCommand = origRun
		runCommandWithStdin = origRunWithStdin
		readFile = origRead
		writeFile = origWrite
		validateSudoersContent = origValidate
		sudoersDPath = origSudoersD
	})

	lookupUserByName = func(name string) (any, error) { return nil, errors.New("not found") }
	runCommand = func(name string, args ...string) error { return nil }
	runCommandWithStdin = func(name string, stdin string, args ...string) error { return nil }
	readFile = func(path string) ([]byte, error) { return nil, fs.ErrNotExist }

	wrote := false
	writeFile = func(path string, data []byte, perm fs.FileMode) error { wrote = true; return nil }
	validateSudoersContent = func(content string) error { return errors.New("invalid sudoers") }

	sudoersDPath = "/fake/etc/sudoers.d"

	if err := createUser("erin", "pw"); err == nil {
		t.Fatalf("expected error when sudoers validation fails")
	}
	if wrote {
		t.Fatalf("did not expect sudoers file to be written when validation fails")
	}
}
