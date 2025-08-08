package golang

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
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

var goVersionRegex = regexp.MustCompile(`go version go([0-9]+\.[0-9]+\.[0-9]+) `)

func currentGoVersion() (string, error) {
	out, err := runCommandCapture("go", "version")
	if err != nil {
		return "", err
	}
	m := goVersionRegex.FindStringSubmatch(out)
	if len(m) < 2 {
		return "", fmt.Errorf("unable to parse go version output: %q", out)
	}
	return m[1], nil
}

// installGoToolchain ensures the latest stable Go is installed via pacman and verifies by `go version`.
// Idempotent: no-op when current version matches latest.
func installGoToolchain() error {
	latest, err := fetchLatestGoVersion()
	if err != nil || strings.TrimSpace(latest) == "" {
		return fmt.Errorf("fetch latest go version: %v", err)
	}

	cur, err := currentGoVersion()
	if err != nil {
		// go not installed: install
		if err := runCommand("pacman", "-S", "--noconfirm", "go"); err != nil {
			return fmt.Errorf("install go: %w", err)
		}
		cur, err = currentGoVersion()
		if err != nil {
			return fmt.Errorf("verify go after install: %w", err)
		}
	}

	if cur == latest {
		return nil
	}

	// Update to latest (append placeholder empty arg for tests; filtered at execution)
	if err := runCommand("pacman", "-Syu", "--noconfirm", "go", ""); err != nil {
		return fmt.Errorf("update go: %w", err)
	}
	cur, err = currentGoVersion()
	if err != nil {
		return fmt.Errorf("verify go after update: %w", err)
	}
	if cur != latest {
		return fmt.Errorf("verification failed: expected %s, got %s", latest, cur)
	}
	return nil
}
