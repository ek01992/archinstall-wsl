package golang

import (
	"context"
	"errors"
	"os/exec"
	"time"
)

func filterEmpty(args []string) []string {
	out := make([]string, 0, len(args))
	for _, a := range args {
		if a != "" {
			out = append(out, a)
		}
	}
	return out
}

// NOTE: Package-level seams are for testability and are NOT concurrency-safe.
// Use internal/seams.With in tests to serialize overrides. Prefer DI if adding concurrency.
var (
	runCommand = func(name string, args ...string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		cmd := exec.CommandContext(ctx, name, filterEmpty(args)...)
		return cmd.Run()
	}
	runCommandCapture = func(name string, args ...string) (string, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
		defer cancel()
		cmd := exec.CommandContext(ctx, name, filterEmpty(args)...)
		out, err := cmd.Output()
		return string(out), err
	}
	fetchLatestGoVersion = func() (string, error) {
		// In real code we'd query upstream, but tests stub this.
		return "", errors.New("not implemented")
	}
)

type seamRunner struct{}

func (seamRunner) Run(name string, args ...string) error { return runCommand(name, args...) }
func (seamRunner) Output(name string, args ...string) (string, error) {
	return runCommandCapture(name, args...)
}

type seamVS struct{}

func (seamVS) LatestGo() (string, error) { return fetchLatestGoVersion() }

// installGoToolchain ensures the latest stable Go is installed via pacman and verifies by `go version`.
// Idempotent: no-op when current version matches latest.
func installGoToolchain() error { return NewService(seamRunner{}, seamVS{}).Install() }
