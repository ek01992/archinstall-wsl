package user

// Legacy tests: to be removed after DI migration.

import (
	"errors"
	"testing"
)

func TestInstallZshTx_RollsBackOnFailure(t *testing.T) {
	origGet := getTargetUsername
	origRun := runCommand
	origReadPasswd := readPasswdFileBytes

	t.Cleanup(func() { getTargetUsername = origGet; runCommand = origRun; readPasswdFileBytes = origReadPasswd })

	getTargetUsername = func() string { return "alice" }
	// Simulate current shell /bin/bash
	stateZsh := false
	readPasswdFileBytes = func() ([]byte, error) {
		if stateZsh {
			return []byte("alice:x:1000:1000::/home/alice:/usr/bin/zsh\n"), nil
		}
		return []byte("alice:x:1000:1000::/home/alice:/bin/bash\n"), nil
	}

	// Fail the install step and capture rollback chsh
	rollbackCalled := false
	runCommand = func(name string, args ...string) error {
		if name == "chsh" && len(args) == 3 && args[0] == "-s" && args[1] == "/usr/bin/zsh" {
			return errors.New("fail")
		}
		if name == "chsh" && len(args) == 3 && args[0] == "-s" && args[1] == "/bin/bash" {
			rollbackCalled = true
			stateZsh = false
			return nil
		}
		return nil
	}

	if err := installZshTx(); err == nil {
		t.Fatalf("expected error from installZshTx")
	}
	if !rollbackCalled {
		t.Fatalf("expected rollback to restore /bin/bash")
	}
}

func TestInstallZshTx_SuccessDoesNotRollback(t *testing.T) {
	origGet := getTargetUsername
	origRun := runCommand
	origReadPasswd := readPasswdFileBytes

	t.Cleanup(func() { getTargetUsername = origGet; runCommand = origRun; readPasswdFileBytes = origReadPasswd })

	getTargetUsername = func() string { return "bob" }
	stateZsh := false
	readPasswdFileBytes = func() ([]byte, error) {
		if stateZsh {
			return []byte("bob:x:1001:1001::/home/bob:/usr/bin/zsh\n"), nil
		}
		return []byte("bob:x:1001:1001::/home/bob:/bin/bash\n"), nil
	}

	rollbackCalled := false
	runCommand = func(name string, args ...string) error {
		if name == "chsh" && len(args) == 3 && args[0] == "-s" && args[1] == "/usr/bin/zsh" && args[2] == "bob" {
			stateZsh = true
			return nil
		}
		if name == "chsh" && len(args) == 3 && args[0] == "-s" && args[1] == "/bin/bash" && args[2] == "bob" {
			rollbackCalled = true
		}
		return nil
	}

	if err := installZshTx(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rollbackCalled {
		t.Fatalf("did not expect rollback on success")
	}
}
