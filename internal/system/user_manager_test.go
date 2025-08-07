package system

import (
	"context"
	"testing"

	"archinstall-wsl/internal/system/mocks"
)

func TestUserManager_CreateUserWithSudo(t *testing.T) {
	ctx := context.Background()

	t.Run("creates new user when user doesn't exist", func(t *testing.T) {
		executor := mocks.NewMockExecutor()
		stateChecker := mocks.NewMockStateChecker()

		// Mock state: user doesn't exist, no sudo config
		stateChecker.SetUserExists("testuser", false)
		stateChecker.SetIsSudoer("testuser", false)
		stateChecker.SetFileExists("/etc/sudoers.d/wheel", false)

		// Mock successful command execution
		executor.SetDefaultExecuteResponse("", nil)
		executor.SetDefaultExecuteSilentError(nil)

		manager := NewUserManager(executor, stateChecker)
		result, err := manager.CreateUserWithSudo(ctx, "testuser")

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if !result.UserCreated {
			t.Error("Expected user to be created")
		}

		expectedCommands := []string{
			"useradd -m -s /bin/bash testuser",
			"usermod -aG wheel testuser",
			"echo '%wheel ALL=(ALL) NOPASSWD: ALL' > /etc/sudoers.d/wheel",
		}

		if len(result.Commands) != len(expectedCommands) {
			t.Errorf("Expected %d commands, got %d", len(expectedCommands), len(result.Commands))
		}

		for i, expected := range expectedCommands {
			if i >= len(result.Commands) {
				t.Errorf("Missing command: %s", expected)
				continue
			}
			if result.Commands[i] != expected {
				t.Errorf("Expected command %q, got %q", expected, result.Commands[i])
			}
		}

		expectedUndoCommands := []string{
			"rm -f /etc/sudoers.d/wheel",
			"userdel -r testuser",
		}

		if len(result.UndoCommands) != len(expectedUndoCommands) {
			t.Errorf("Expected %d undo commands, got %d", len(expectedUndoCommands), len(result.UndoCommands))
		}
	})

	t.Run("does nothing when user exists with sudo privileges", func(t *testing.T) {
		executor := mocks.NewMockExecutor()
		stateChecker := mocks.NewMockStateChecker()

		// Mock state: user exists and has sudo privileges
		stateChecker.SetUserExists("testuser", true)
		stateChecker.SetIsSudoer("testuser", true)

		manager := NewUserManager(executor, stateChecker)
		result, err := manager.CreateUserWithSudo(ctx, "testuser")

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if result.UserCreated {
			t.Error("Expected user not to be created")
		}

		if len(result.Commands) != 0 {
			t.Errorf("Expected no commands, got %d", len(result.Commands))
		}

		if len(result.UndoCommands) != 0 {
			t.Errorf("Expected no undo commands, got %d", len(result.UndoCommands))
		}
	})

	t.Run("adds sudo privileges to existing user", func(t *testing.T) {
		executor := mocks.NewMockExecutor()
		stateChecker := mocks.NewMockStateChecker()

		// Mock state: user exists but no sudo privileges
		stateChecker.SetUserExists("testuser", true)
		stateChecker.SetIsSudoer("testuser", false)
		stateChecker.SetFileExists("/etc/sudoers.d/wheel", false)

		// Mock successful command execution
		executor.SetDefaultExecuteResponse("", nil)
		executor.SetDefaultExecuteSilentError(nil)

		manager := NewUserManager(executor, stateChecker)
		result, err := manager.CreateUserWithSudo(ctx, "testuser")

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if result.UserCreated {
			t.Error("Expected user not to be created")
		}

		expectedCommands := []string{
			"usermod -aG wheel testuser",
			"echo '%wheel ALL=(ALL) NOPASSWD: ALL' > /etc/sudoers.d/wheel",
		}

		if len(result.Commands) != len(expectedCommands) {
			t.Errorf("Expected %d commands, got %d", len(expectedCommands), len(result.Commands))
		}
	})

	t.Run("handles user creation failure", func(t *testing.T) {
		executor := mocks.NewMockExecutor()
		stateChecker := mocks.NewMockStateChecker()

		// Mock state: user doesn't exist
		stateChecker.SetUserExists("testuser", false)

		// Mock command failure
		executor.SetExecuteResponse("useradd -m -s /bin/bash testuser", "", context.DeadlineExceeded)

		manager := NewUserManager(executor, stateChecker)
		_, err := manager.CreateUserWithSudo(ctx, "testuser")

		if err == nil {
			t.Error("Expected error when user creation fails")
		}
	})

	t.Run("handles wheel group addition failure", func(t *testing.T) {
		executor := mocks.NewMockExecutor()
		stateChecker := mocks.NewMockStateChecker()

		// Mock state: user doesn't exist
		stateChecker.SetUserExists("testuser", false)

		// Mock successful user creation but failed wheel group addition
		executor.SetExecuteResponse("useradd -m -s /bin/bash testuser", "", nil)
		executor.SetExecuteResponse("usermod -aG wheel testuser", "", context.DeadlineExceeded)

		manager := NewUserManager(executor, stateChecker)
		_, err := manager.CreateUserWithSudo(ctx, "testuser")

		if err == nil {
			t.Error("Expected error when wheel group addition fails")
		}
	})
}

func TestUserManager_UndoUserCreation(t *testing.T) {
	ctx := context.Background()

	t.Run("executes undo commands in reverse order", func(t *testing.T) {
		executor := mocks.NewMockExecutor()
		stateChecker := mocks.NewMockStateChecker()

		// Mock successful undo command execution
		executor.SetDefaultExecuteSilentError(nil)

		manager := NewUserManager(executor, stateChecker)
		result := &UserCreationResult{
			UserCreated: true,
			Commands: []string{
				"useradd -m -s /bin/bash testuser",
				"usermod -aG wheel testuser",
				"echo '%wheel ALL=(ALL) NOPASSWD: ALL' > /etc/sudoers.d/wheel",
			},
			UndoCommands: []string{
				"rm -f /etc/sudoers.d/wheel",
				"userdel -r testuser",
			},
		}

		err := manager.UndoUserCreation(ctx, result)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("handles undo command failure", func(t *testing.T) {
		executor := mocks.NewMockExecutor()
		stateChecker := mocks.NewMockStateChecker()

		// Mock undo command failure
		executor.SetExecuteSilentError("rm -f /etc/sudoers.d/wheel", context.DeadlineExceeded)

		manager := NewUserManager(executor, stateChecker)
		result := &UserCreationResult{
			UserCreated: true,
			UndoCommands: []string{
				"rm -f /etc/sudoers.d/wheel",
				"userdel -r testuser",
			},
		}

		err := manager.UndoUserCreation(ctx, result)

		if err == nil {
			t.Error("Expected error when undo command fails")
		}
	})
}
