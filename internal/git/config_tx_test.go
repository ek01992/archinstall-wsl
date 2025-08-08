package git

import (
	"strings"
	"testing"
)

func TestConfigureGitTx_RollsBackOnFailure(t *testing.T) {
	origRun := runCommand
	origCap := runCommandCapture
	t.Cleanup(func() { runCommand = origRun; runCommandCapture = origCap })

	// Prior values
	runCommandCapture = func(name string, args ...string) (string, error) {
		if len(args) >= 1 && args[0] == "config" && len(args) >= 4 && args[2] == "--get" {
			if args[3] == "user.name" {
				return "Old Name\n", nil
			}
			if args[3] == "user.email" {
				return "old@example.com\n", nil
			}
		}
		return "", nil
	}

	var calls []string
	runCommand = func(name string, args ...string) error {
		calls = append(calls, strings.Join(append([]string{name}, args...), " "))
		// Fail during configureGit set call
		if name == "git" && len(args) >= 4 && args[0] == "config" && args[2] == "user.name" {
			return assertErr
		}
		return nil
	}

	if err := configureGitTx("New Name", "new@example.com"); err == nil {
		t.Fatalf("expected error")
	}
	// Rollback should attempt to restore old values
	foundName := false
	foundEmail := false
	for _, c := range calls {
		if strings.Contains(c, "git config --global user.name Old Name") {
			foundName = true
		}
		if strings.Contains(c, "git config --global user.email old@example.com") {
			foundEmail = true
		}
	}
	if !foundName || !foundEmail {
		t.Fatalf("expected rollback to restore both values; calls=%v", calls)
	}
}

func TestConfigureGitTx_SuccessDoesNotRollback(t *testing.T) {
	origRun := runCommand
	origCap := runCommandCapture
	t.Cleanup(func() { runCommand = origRun; runCommandCapture = origCap })

	var curName, curEmail string
	// After set, verification should return the new values
	runCommandCapture = func(name string, args ...string) (string, error) {
		if len(args) >= 1 && args[0] == "config" && len(args) >= 4 && args[2] == "--get" {
			if args[3] == "user.name" {
				return curName + "\n", nil
			}
			if args[3] == "user.email" {
				return curEmail + "\n", nil
			}
		}
		return "", nil
	}

	rolledBack := false
	runCommand = func(name string, args ...string) error {
		if name == "git" && len(args) >= 4 && args[0] == "config" && args[2] == "user.name" {
			curName = args[3]
			return nil
		}
		if name == "git" && len(args) >= 4 && args[0] == "config" && args[2] == "user.email" {
			curEmail = args[3]
			return nil
		}
		if len(args) >= 1 && args[0] == "config" && (args[2] == "--unset" || args[2] == "user.name" || args[2] == "user.email") {
			rolledBack = true
		}
		return nil
	}

	if err := configureGitTx("New Name", "new@example.com"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rolledBack {
		t.Fatalf("did not expect rollback on success")
	}
}

type errString string

func (e errString) Error() string { return string(e) }

const assertErr = errString("boom")
