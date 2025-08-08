package user

import (
	"errors"
	"strings"
	"testing"
)

func TestInstallZsh_UsesChshAndVerifies(t *testing.T) {
	origGetTarget := getTargetUsername
	origRun := runCommand
	origReadPasswd := readPasswdFileBytes

	t.Cleanup(func() {
		getTargetUsername = origGetTarget
		runCommand = origRun
		readPasswdFileBytes = origReadPasswd
	})

	getTargetUsername = func() string { return "alice" }

	changed := false
	runCommand = func(name string, args ...string) error {
		if name != "chsh" {
			t.Fatalf("expected chsh, got %q", name)
		}
		if len(args) != 3 || args[0] != "-s" || args[1] != "/usr/bin/zsh" || args[2] != "alice" {
			t.Fatalf("unexpected chsh args: %v", args)
		}
		changed = true
		return nil
	}

	readPasswdFileBytes = func() ([]byte, error) {
		if !changed {
			return []byte("alice:x:1000:1000::/home/alice:/bin/bash\n"), nil
		}
		return []byte("alice:x:1000:1000::/home/alice:/usr/bin/zsh\n"), nil
	}

	if err := installZsh(); err != nil {
		t.Fatalf("installZsh returned error: %v", err)
	}
	if !changed {
		t.Fatalf("expected chsh to be called")
	}
}

func TestInstallZsh_FallbackToUsermodOnChshFailure(t *testing.T) {
	origGetTarget := getTargetUsername
	origRun := runCommand
	origReadPasswd := readPasswdFileBytes

	t.Cleanup(func() {
		getTargetUsername = origGetTarget
		runCommand = origRun
		readPasswdFileBytes = origReadPasswd
	})

	getTargetUsername = func() string { return "bob" }

	calledChsh := false
	calledUsermod := false
	runCommand = func(name string, args ...string) error {
		switch name {
		case "chsh":
			calledChsh = true
			return errors.New("chsh not available")
		case "usermod":
			if len(args) != 3 || args[0] != "-s" || args[1] != "/usr/bin/zsh" || args[2] != "bob" {
				t.Fatalf("unexpected usermod args: %v", args)
			}
			calledUsermod = true
			return nil
		default:
			t.Fatalf("unexpected command %q", name)
			return nil
		}
	}

	changed := false
	readPasswdFileBytes = func() ([]byte, error) {
		if !changed {
			// Flip after first read to simulate change after usermod runs
			if calledUsermod {
				changed = true
			}
			return []byte("bob:x:1001:1001::/home/bob:/bin/bash\n"), nil
		}
		return []byte("bob:x:1001:1001::/home/bob:/usr/bin/zsh\n"), nil
	}

	if err := installZsh(); err != nil {
		t.Fatalf("installZsh returned error: %v", err)
	}
	if !calledChsh || !calledUsermod {
		t.Fatalf("expected chsh then usermod to be called; got chsh=%v usermod=%v", calledChsh, calledUsermod)
	}
}

func TestInstallZsh_IdempotentWhenAlreadyZsh(t *testing.T) {
	origGetTarget := getTargetUsername
	origRun := runCommand
	origReadPasswd := readPasswdFileBytes

	t.Cleanup(func() {
		getTargetUsername = origGetTarget
		runCommand = origRun
		readPasswdFileBytes = origReadPasswd
	})

	getTargetUsername = func() string { return "carol" }

	readPasswdFileBytes = func() ([]byte, error) {
		return []byte("carol:x:1002:1002::/home/carol:/usr/bin/zsh\n"), nil
	}

	runCommand = func(name string, args ...string) error {
		t.Fatalf("no commands should be run when shell already zsh; got %q %v", name, args)
		return nil
	}

	if err := installZsh(); err != nil {
		t.Fatalf("installZsh returned error: %v", err)
	}
}

func TestInstallZsh_EmptyTargetUserReturnsError(t *testing.T) {
	origGetTarget := getTargetUsername
	origRun := runCommand
	origReadPasswd := readPasswdFileBytes

	t.Cleanup(func() {
		getTargetUsername = origGetTarget
		runCommand = origRun
		readPasswdFileBytes = origReadPasswd
	})

	getTargetUsername = func() string { return "  \t\n" }

	runCommand = func(name string, args ...string) error {
		t.Fatalf("no commands should be run for empty target user")
		return nil
	}
	readPasswdFileBytes = func() ([]byte, error) {
		t.Fatalf("passwd should not be read for empty target user")
		return nil, nil
	}

	if err := installZsh(); err == nil {
		t.Fatalf("expected error for empty target user")
	} else if !strings.Contains(err.Error(), "empty target user") {
		t.Fatalf("unexpected error: %v", err)
	}
}
