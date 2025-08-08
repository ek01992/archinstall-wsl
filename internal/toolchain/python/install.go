package python

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// NOTE: Package-level seams are for testability and are NOT concurrency-safe.
// Use internal/seams.With in tests to serialize overrides. Prefer DI if adding concurrency.
var (
	runCommand = func(name string, args ...string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		cmd := exec.CommandContext(ctx, name, args...)
		return cmd.Run()
	}
	runCommandCapture = func(name string, args ...string) (string, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
		defer cancel()
		cmd := exec.CommandContext(ctx, name, args...)
		out, err := cmd.CombinedOutput()
		return string(out), err
	}
	fetchLatestPythonVersion = func() (string, error) {
		return "", errors.New("not implemented")
	}
)

var pythonVersionRegex = regexp.MustCompile(`(?i)^Python\s+([0-9]+\.[0-9]+\.[0-9]+)`) // matches "Python 3.x.y"

func currentPythonVersion() (string, error) {
	out, err := runCommandCapture("python", "--version")
	if err != nil {
		return "", err
	}
	m := pythonVersionRegex.FindStringSubmatch(strings.TrimSpace(out))
	if len(m) < 2 {
		return "", fmt.Errorf("unable to parse python version output: %q", out)
	}
	return m[1], nil
}

func isPyenvInstalled() bool {
	_, err := runCommandCapture("pyenv", "--version")
	return err == nil
}

func ensurePyenvInstalled() error {
	if isPyenvInstalled() {
		return nil
	}
	if err := runCommand("pacman", "-S", "--noconfirm", "pyenv"); err != nil {
		return fmt.Errorf("install pyenv: %w", err)
	}
	return nil
}

func ensurePythonVersion(version string) error {
	// Install if missing (-s skips if already installed)
	if err := runCommand("pyenv", "install", "-s", version); err != nil {
		// Some stubs may pass combined "-s <ver>"; our tests accept either; we keep separate args here
		return fmt.Errorf("pyenv install %s: %w", version, err)
	}
	// Set global version
	if err := runCommand("pyenv", "global", version); err != nil {
		return fmt.Errorf("pyenv global %s: %w", version, err)
	}
	return nil
}

func isPipxInstalled() bool {
	_, err := runCommandCapture("pipx", "--version")
	return err == nil
}

// installPythonToolchain installs/updates Python via pyenv and ensures pipx is present. Idempotent.
func installPythonToolchain() error {
	latest, err := fetchLatestPythonVersion()
	if err != nil || strings.TrimSpace(latest) == "" {
		return fmt.Errorf("fetch latest python: %w", err)
	}

	if err := ensurePyenvInstalled(); err != nil {
		return err
	}

	cur, err := currentPythonVersion()
	if err != nil || cur != latest {
		if err := ensurePythonVersion(latest); err != nil {
			return err
		}
		// Verify
		cur2, err := currentPythonVersion()
		if err != nil {
			return fmt.Errorf("verify python after configure: %w", err)
		}
		if cur2 != latest {
			return fmt.Errorf("verification failed: expected Python %s, got %s", latest, cur2)
		}
	}

	if !isPipxInstalled() {
		if err := runCommand("pacman", "-S", "--noconfirm", "pipx"); err != nil {
			return fmt.Errorf("install pipx: %w", err)
		}
		// optional verify
		if !isPipxInstalled() {
			return fmt.Errorf("verification failed: pipx not available")
		}
	}

	return nil
}
