package nodejs

import (
	"context"
	"errors"
	"os/exec"
	"regexp"
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
	fetchLatestNodeLTS = func() (string, error) { return "", errors.New("not implemented") }
)

var nodeVersionRegex = regexp.MustCompile(`^v\d+\.\d+\.\d+$`)

type seamRunner struct{}
func (seamRunner) Run(name string, args ...string) error            { return runCommand(name, args...) }
func (seamRunner) Output(name string, args ...string) (string, error) { return runCommandCapture(name, args...) }
func (seamRunner) Shell(cmd string) (string, error)                 { return runShellCapture(cmd) }

type seamVS struct{}
func (seamVS) LatestLTS() (string, error) { return fetchLatestNodeLTS() }

func runShell(cmd string) error { return exec.Command("bash", "-lc", cmd).Run() }
func runShellCapture(cmd string) (string, error) {
	out, err := exec.Command("bash", "-lc", cmd).CombinedOutput()
	return string(out), err
}

// installNodeToolchain installs nvm if missing, ensures latest LTS is installed and default-selected.
func installNodeToolchain() error { return NewService(seamRunner{}, seamVS{}).Install() }
