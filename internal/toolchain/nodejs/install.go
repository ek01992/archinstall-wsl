package nodejs

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

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

func isNvmInstalled() bool {
	if _, err := runCommandCapture("nvm", "--version"); err == nil {
		return true
	}
	// Fallback via bash -lc in case nvm is a shell function
	if _, err := runShellCapture("nvm --version"); err == nil {
		return true
	}
	return false
}

func ensureNvmInstalled() error {
	if isNvmInstalled() {
		return nil
	}
	if err := runCommand("pacman", "-S", "--noconfirm", "nvm"); err != nil {
		return fmt.Errorf("install nvm: %w", err)
	}
	return nil
}

func currentNodeVersion() (string, error) {
	out, err := runCommandCapture("node", "-v")
	if err != nil {
		// Try via shell in case node is in a shell-managed PATH
		if out2, err2 := runShellCapture("node -v"); err2 == nil {
			out = out2
		} else {
			return "", err
		}
	}
	s := strings.TrimSpace(out)
	if !nodeVersionRegex.MatchString(s) {
		return "", fmt.Errorf("unable to parse node -v output: %q", s)
	}
	return s, nil
}

func runShell(cmd string) error {
	return exec.Command("bash", "-lc", cmd).Run()
}

func runShellCapture(cmd string) (string, error) {
	out, err := exec.Command("bash", "-lc", cmd).CombinedOutput()
	return string(out), err
}

// installNodeToolchain installs nvm if missing, ensures latest LTS is installed and default-selected.
func installNodeToolchain() error {
	lts, err := fetchLatestNodeLTS()
	if err != nil || strings.TrimSpace(lts) == "" {
		return fmt.Errorf("fetch latest node LTS: %v", err)
	}

	if err := ensureNvmInstalled(); err != nil {
		return err
	}

	cur, err := currentNodeVersion()
	if err != nil || cur != lts {
		// Try direct; if it fails, try shell-based invocation
		if err := runCommand("nvm", "install", lts); err != nil {
			if err2 := runShell("source /usr/share/nvm/init-nvm.sh 2>/dev/null || true; nvm install "+lts); err2 != nil {
				return fmt.Errorf("nvm install %s: %w", lts, err)
			}
		}
		if err := runCommand("nvm", "alias", "default", lts); err != nil {
			if err2 := runShell("source /usr/share/nvm/init-nvm.sh 2>/dev/null || true; nvm alias default "+lts); err2 != nil {
				return fmt.Errorf("nvm alias default %s: %w", lts, err)
			}
		}
		nv, err := currentNodeVersion()
		if err != nil {
			return fmt.Errorf("verify node after install: %w", err)
		}
		if nv != lts {
			return fmt.Errorf("verification failed: expected %s, got %s", lts, nv)
		}
	}

	return nil
}
