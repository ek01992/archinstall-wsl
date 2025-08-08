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
    _, err := runCommandCapture("nvm", "--version")
    return err == nil
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
        return "", err
    }
    s := strings.TrimSpace(out)
    if !nodeVersionRegex.MatchString(s) {
        return "", fmt.Errorf("unable to parse node -v output: %q", s)
    }
    return s, nil
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
        if err := runCommand("nvm", "install", lts); err != nil {
            return fmt.Errorf("nvm install %s: %w", lts, err)
        }
        if err := runCommand("nvm", "alias", "default", lts); err != nil {
            return fmt.Errorf("nvm alias default %s: %w", lts, err)
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
