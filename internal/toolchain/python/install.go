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

type seamRunner struct{}

func (seamRunner) Run(name string, args ...string) error { return runCommand(name, args...) }
func (seamRunner) Output(name string, args ...string) (string, error) {
	return runCommandCapture(name, args...)
}

type seamVS struct{}

func (seamVS) LatestPython() (string, error) { return fetchLatestPythonVersion() }

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
func installPythonToolchain() error { return NewService(seamRunner{}, seamVS{}).Install() }
