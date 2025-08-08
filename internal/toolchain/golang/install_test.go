package golang

import (
	"errors"
	"strings"
	"testing"
)

func TestInstallGoToolchain_InstallsWhenMissing(t *testing.T) {
	origFetch := fetchLatestGoVersion
	origRun := runCommand
	origCap := runCommandCapture
	t.Cleanup(func() {
		fetchLatestGoVersion = origFetch
		runCommand = origRun
		runCommandCapture = origCap
	})

	fetchLatestGoVersion = func() (string, error) { return "1.20.0", nil }

	installed := false
	runCommand = func(name string, args ...string) error {
		if name != "pacman" {
			t.Fatalf("expected pacman, got %q", name)
		}
		// Expect initial install when missing
		joined := strings.Join(args, " ")
		if strings.HasPrefix(joined, "-S --noconfirm go") {
			installed = true
			return nil
		}
		t.Fatalf("unexpected pacman args: %v", args)
		return nil
	}

	runCommandCapture = func(name string, args ...string) (string, error) {
		if name != "go" || len(args) != 1 || args[0] != "version" {
			t.Fatalf("unexpected capture call: %q %v", name, args)
		}
		if !installed {
			return "", errors.New("go not found")
		}
		return "go version go1.20.0 linux/amd64", nil
	}

	if err := installGoToolchain(); err != nil {
		t.Fatalf("installGoToolchain returned error: %v", err)
	}
	if !installed {
		t.Fatalf("expected pacman install to be called")
	}
}

func TestInstallGoToolchain_UpdatesWhenOutdated(t *testing.T) {
	origFetch := fetchLatestGoVersion
	origRun := runCommand
	origCap := runCommandCapture
	t.Cleanup(func() {
		fetchLatestGoVersion = origFetch
		runCommand = origRun
		runCommandCapture = origCap
	})

	fetchLatestGoVersion = func() (string, error) { return "1.20.0", nil }

	updated := false
	runCommand = func(name string, args ...string) error {
		if name != "pacman" {
			t.Fatalf("expected pacman, got %q", name)
		}
		joined := strings.Join(args, " ")
		if strings.HasPrefix(joined, "-Syu --noconfirm go") {
			updated = true
			return nil
		}
		t.Fatalf("unexpected pacman args: %v", args)
		return nil
	}

	// First returns old version, then new after update
	returnedNew := false
	runCommandCapture = func(name string, args ...string) (string, error) {
		if name != "go" || len(args) != 1 || args[0] != "version" {
			t.Fatalf("unexpected capture call: %q %v", name, args)
		}
		if !returnedNew {
			returnedNew = true
			return "go version go1.19.0 linux/amd64", nil
		}
		return "go version go1.20.0 linux/amd64", nil
	}

	if err := installGoToolchain(); err != nil {
		t.Fatalf("installGoToolchain returned error: %v", err)
	}
	if !updated {
		t.Fatalf("expected pacman update to be called")
	}
}

func TestInstallGoToolchain_IdempotentWhenUpToDate(t *testing.T) {
	origFetch := fetchLatestGoVersion
	origRun := runCommand
	origCap := runCommandCapture
	t.Cleanup(func() {
		fetchLatestGoVersion = origFetch
		runCommand = origRun
		runCommandCapture = origCap
	})

	fetchLatestGoVersion = func() (string, error) { return "1.20.0", nil }

	runCommand = func(name string, args ...string) error {
		t.Fatalf("no pacman calls expected when up-to-date, got %q %v", name, args)
		return nil
	}

	runCommandCapture = func(name string, args ...string) (string, error) {
		return "go version go1.20.0 linux/amd64", nil
	}

	if err := installGoToolchain(); err != nil {
		t.Fatalf("installGoToolchain returned error: %v", err)
	}
}

func TestInstallGoToolchain_FetchFails(t *testing.T) {
	origFetch := fetchLatestGoVersion
	origRun := runCommand
	t.Cleanup(func() { fetchLatestGoVersion = origFetch; runCommand = origRun })

	fetchLatestGoVersion = func() (string, error) { return "", errors.New("fail") }

	runCommand = func(name string, args ...string) error {
		t.Fatalf("no pacman calls expected when fetch fails")
		return nil
	}

	if err := installGoToolchain(); err == nil {
		t.Fatalf("expected error when fetching latest version fails")
	}
}
