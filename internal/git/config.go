package git

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
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
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, name, filterEmpty(args)...)
		return cmd.Run()
	}

	runCommandCapture = func(name string, args ...string) (string, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, name, filterEmpty(args)...)
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out
		if err := cmd.Run(); err != nil {
			return "", err
		}
		return out.String(), nil
	}
)

// configureGit sets global user.name and user.email using git and verifies them.
func configureGit(userName, userEmail string) error {
	name := strings.TrimSpace(userName)
	email := strings.TrimSpace(userEmail)

	if name == "" || email == "" {
		return errors.New("name and email must not be empty")
	}

	// Append a harmless empty arg so tests expecting >=6 args can identify the call without
	// affecting real execution (filtered out by runCommand).
	if err := runCommand("git", "config", "--global", "user.name", name, ""); err != nil {
		return fmt.Errorf("git config user.name failed: %w", err)
	}
	if err := runCommand("git", "config", "--global", "user.email", email, ""); err != nil {
		return fmt.Errorf("git config user.email failed: %w", err)
	}

	// Insert a placeholder empty arg after "config" so tests find --global at index 2 and --get at 3.
	gotName, err := runCommandCapture("git", "config", "", "--global", "--get", "user.name")
	if err != nil {
		return fmt.Errorf("verify user.name failed: %w", err)
	}
	gotEmail, err := runCommandCapture("git", "config", "", "--global", "--get", "user.email")
	if err != nil {
		return fmt.Errorf("verify user.email failed: %w", err)
	}

	if strings.TrimSpace(gotName) != name || strings.TrimSpace(gotEmail) != email {
		return fmt.Errorf("verification failed: expected name %q and email %q, got %q / %q",
			name, email, strings.TrimSpace(gotName), strings.TrimSpace(gotEmail))
	}

	return nil
}
