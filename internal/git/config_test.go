package git

// Legacy tests: to be removed after DI migration.

import (
	"errors"
	"strings"
	"testing"
)

func TestConfigureGit_SetsAndVerifies(t *testing.T) {
	origRun := runCommand
	origRunCap := runCommandCapture
	t.Cleanup(func() {
		runCommand = origRun
		runCommandCapture = origRunCap
	})

	var calls [][]string
	runCommand = func(name string, args ...string) error {
		if name != "git" {
			t.Fatalf("expected git command, got %q", name)
		}
		calls = append(calls, append([]string{name}, args...))
		return nil
	}

	runCommandCapture = func(name string, args ...string) (string, error) {
		if name != "git" {
			t.Fatalf("expected git command, got %q", name)
		}
		joined := strings.Join(args, " ")
		if strings.HasPrefix(joined, "config --global --get user.name") {
			return "Alice Example\n", nil
		}
		if strings.HasPrefix(joined, "config --global --get user.email") {
			return "alice@example.com\n", nil
		}
		return "", errors.New("unexpected capture call")
	}

	if err := configureGit(" Alice Example ", " alice@example.com "); err != nil {
		t.Fatalf("configureGit returned error: %v", err)
	}

	// Validate set calls were issued exactly for name and email
	var sawName, sawEmail bool
	for _, c := range calls {
		args := strings.Join(c[1:], " ")
		if strings.HasPrefix(args, "config --global user.name Alice Example") {
			sawName = true
		}
		if strings.HasPrefix(args, "config --global user.email alice@example.com") {
			sawEmail = true
		}
	}
	if !sawName {
		t.Fatalf("expected git config --global user.name to be called with trimmed value; calls=%v", calls)
	}
	if !sawEmail {
		t.Fatalf("expected git config --global user.email to be called with trimmed value; calls=%v", calls)
	}
}

func TestConfigureGit_EmptyInputsReturnError_NoCommandsRun(t *testing.T) {
	origRun := runCommand
	origRunCap := runCommandCapture
	t.Cleanup(func() {
		runCommand = origRun
		runCommandCapture = origRunCap
	})

	runCommand = func(name string, args ...string) error {
		t.Fatalf("no commands should be run for empty inputs")
		return nil
	}
	runCommandCapture = func(name string, args ...string) (string, error) {
		t.Fatalf("no capture commands should be run for empty inputs")
		return "", nil
	}

	if err := configureGit(" \t\n", "alice@example.com"); err == nil {
		t.Fatalf("expected error for empty name")
	}
	if err := configureGit("Alice", "  \t\n"); err == nil {
		t.Fatalf("expected error for empty email")
	}
}

func TestConfigureGit_VerificationFailureReturnsError(t *testing.T) {
	origRun := runCommand
	origRunCap := runCommandCapture
	t.Cleanup(func() {
		runCommand = origRun
		runCommandCapture = origRunCap
	})

	runCommand = func(name string, args ...string) error { return nil }
	runCommandCapture = func(name string, args ...string) (string, error) {
		joined := strings.Join(args, " ")
		if strings.HasPrefix(joined, "config --global --get user.name") {
			return "Wrong Name\n", nil
		}
		if strings.HasPrefix(joined, "config --global --get user.email") {
			return "wrong@example.com\n", nil
		}
		return "", nil
	}

	err := configureGit("Alice", "alice@example.com")
	if err == nil {
		t.Fatalf("expected verification error, got nil")
	}
	if !strings.Contains(err.Error(), "verification failed") {
		t.Fatalf("expected verification failure message, got %v", err)
	}
}
