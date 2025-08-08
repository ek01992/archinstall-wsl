package user

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Command and file operation seams for testability.
// NOTE: These globals are NOT concurrency-safe. Tests must use internal/seams.With
// to serialize overrides. Prefer dependency injection for concurrency-sensitive code.
var (
	runCommand = func(name string, args ...string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, name, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	runCommandWithStdin = func(name string, stdin string, args ...string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, name, args...)
		cmd.Stdin = strings.NewReader(stdin)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	readFile = func(path string) ([]byte, error) { return os.ReadFile(path) }

	writeFile = func(path string, data []byte, perm fs.FileMode) error { return os.WriteFile(path, data, perm) }

	mkdirAll = func(path string, perm fs.FileMode) error { return os.MkdirAll(path, perm) }

	sudoersDPath = "/etc/sudoers.d"

	validateSudoersContent = func(content string) error { return nil }
)

// createUser creates the user if missing, sets the password, ensures membership in
// the wheel group, and enables passwordless sudo for wheel via sudoers.d. It is
// idempotent: safe to call repeatedly.
func createUser(username, password string) error {
	username = strings.TrimSpace(username)
	if username == "" {
		return errors.New("username must not be empty")
	}
	// Basic credential sanitation to avoid malformed input to chpasswd
	if strings.ContainsAny(username, ":\n") || strings.Contains(password, "\n") {
		return fmt.Errorf("invalid credentials")
	}

	if !doesUserExist(username) {
		if err := runCommand("useradd", "-m", username); err != nil {
			return fmt.Errorf("useradd failed: %w", err)
		}
	}

	// Set or reset the password (empty password allowed if caller desires)
	if err := runCommandWithStdin("chpasswd", fmt.Sprintf("%s:%s", username, password)); err != nil {
		return fmt.Errorf("setting password failed: %w", err)
	}

	// Ensure wheel group membership. Try usermod first, then gpasswd fallback.
	if err := runCommand("usermod", "-aG", "wheel", username); err != nil {
		if err2 := runCommand("gpasswd", "-a", username, "wheel"); err2 != nil {
			return fmt.Errorf("adding to wheel failed (usermod: %v, gpasswd: %w)", err, err2)
		}
	}

	// Ensure sudoers.d exists and rule for wheel NOPASSWD exists with desired content.
	desired := "%wheel ALL=(ALL) NOPASSWD: ALL\n"
	sudoersFile := filepath.Join(sudoersDPath, "010_wheel_nopasswd")
	_ = mkdirAll(sudoersDPath, 0o755)

	current, err := readFile(sudoersFile)
	// Ignore read errors; we'll overwrite below if missing or mismatched
	_ = err

	if string(current) != desired {
		// Validate content before writing
		if err := validateSudoersContent(desired); err != nil {
			return fmt.Errorf("sudoers validation failed: %w", err)
		}
		if err := writeFile(sudoersFile, []byte(desired), 0o440); err != nil {
			return fmt.Errorf("writing sudoers file failed: %w", err)
		}
	}

	return nil
}
