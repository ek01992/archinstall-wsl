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

// Command and file operation seams for testability
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

	sudoersDPath = "/etc/sudoers.d"
)

// createUser creates the user if missing, sets the password, ensures membership in
// the wheel group, and enables passwordless sudo for wheel via sudoers.d. It is
// idempotent: safe to call repeatedly.
func createUser(username, password string) error {
	username = strings.TrimSpace(username)
	if username == "" {
		return errors.New("username must not be empty")
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

	// Ensure sudoers.d rule for wheel NOPASSWD exists and matches desired content.
	desired := "%wheel ALL=(ALL) NOPASSWD: ALL\n"
	sudoersFile := filepath.Join(sudoersDPath, "010_wheel_nopasswd")

	current, err := readFile(sudoersFile)
	// Ignore read errors; we'll overwrite below if missing or mismatched
	_ = err

	if string(current) != desired {
		if err := writeFile(sudoersFile, []byte(desired), 0o440); err != nil {
			return fmt.Errorf("writing sudoers file failed: %w", err)
		}
	}

	return nil
}
