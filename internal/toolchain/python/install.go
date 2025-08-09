package python

import (
	"context"
	"errors"
	"os/exec"
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
	fetchLatestPythonVersion = func() (string, error) { return "", errors.New("not implemented") }
)

type seamRunner struct{}

func (seamRunner) Run(name string, args ...string) error { return runCommand(name, args...) }
func (seamRunner) Output(name string, args ...string) (string, error) {
	return runCommandCapture(name, args...)
}

type seamVS struct{}

func (seamVS) LatestPython() (string, error) { return fetchLatestPythonVersion() }

// installPythonToolchain installs/updates Python via pyenv and ensures pipx is present. Idempotent.
func installPythonToolchain() error { return NewService(seamRunner{}, seamVS{}).Install() }
