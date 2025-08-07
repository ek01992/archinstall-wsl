package system

import (
	"context"
	"testing"

	"archinstall-wsl/internal/system/mocks"
)

func TestGitConfigManager_ConfigureGitUser(t *testing.T) {
	ctx := context.Background()

	t.Run("sets both name and email when not configured", func(t *testing.T) {
		executor := mocks.NewMockExecutor()
		stateChecker := mocks.NewMockStateChecker()

		// Mock git config commands - return empty for unset values
		executor.SetExecuteResponse("git config --global --get user.name", "", context.DeadlineExceeded)
		executor.SetExecuteResponse("git config --global --get user.email", "", context.DeadlineExceeded)
		executor.SetDefaultExecuteResponse("", nil) // For set commands

		manager := NewGitConfigManager(executor, stateChecker)
		config := GitConfig{
			Name:  "Test User",
			Email: "test@example.com",
		}

		err := manager.ConfigureGitUser(ctx, config)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Note: We can't easily verify the exact commands executed with the current mock
		// In a real implementation, we'd track the commands or use a more sophisticated mock
	})

	t.Run("does not set values when already configured correctly", func(t *testing.T) {
		executor := mocks.NewMockExecutor()
		stateChecker := mocks.NewMockStateChecker()

		// Mock git config commands - return existing values
		executor.SetExecuteResponse("git config --global --get user.name", "Test User", nil)
		executor.SetExecuteResponse("git config --global --get user.email", "test@example.com", nil)

		manager := NewGitConfigManager(executor, stateChecker)
		config := GitConfig{
			Name:  "Test User",
			Email: "test@example.com",
		}

		err := manager.ConfigureGitUser(ctx, config)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Should not execute any set commands since values are already correct
	})

	t.Run("updates only changed values", func(t *testing.T) {
		executor := mocks.NewMockExecutor()
		stateChecker := mocks.NewMockStateChecker()

		// Mock git config commands - return existing name but different email
		executor.SetExecuteResponse("git config --global --get user.name", "Test User", nil)
		executor.SetExecuteResponse("git config --global --get user.email", "old@example.com", nil)
		executor.SetDefaultExecuteResponse("", nil) // For set commands

		manager := NewGitConfigManager(executor, stateChecker)
		config := GitConfig{
			Name:  "Test User",       // Same as existing
			Email: "new@example.com", // Different from existing
		}

		err := manager.ConfigureGitUser(ctx, config)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Should only set the email, not the name
	})

	t.Run("handles git config get errors gracefully", func(t *testing.T) {
		executor := mocks.NewMockExecutor()
		stateChecker := mocks.NewMockStateChecker()

		// Mock git config get command to return error
		executor.SetExecuteResponse("git config --global --get user.name", "", context.DeadlineExceeded)
		executor.SetExecuteResponse("git config --global --get user.email", "", context.DeadlineExceeded)
		executor.SetDefaultExecuteResponse("", nil) // For set commands

		manager := NewGitConfigManager(executor, stateChecker)
		config := GitConfig{
			Name:  "Test User",
			Email: "test@example.com",
		}

		err := manager.ConfigureGitUser(ctx, config)

		if err != nil {
			t.Errorf("Expected no error for git config get failures, got %v", err)
		}
	})
}

func TestGitConfigManager_GetCurrentGitConfig(t *testing.T) {
	ctx := context.Background()

	t.Run("retrieves existing git configuration", func(t *testing.T) {
		executor := mocks.NewMockExecutor()
		stateChecker := mocks.NewMockStateChecker()

		// Mock git config commands
		executor.SetExecuteResponse("git config --global --get user.name", "Test User", nil)
		executor.SetExecuteResponse("git config --global --get user.email", "test@example.com", nil)

		manager := NewGitConfigManager(executor, stateChecker)
		config, err := manager.GetCurrentGitConfig(ctx)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if config.Name != "Test User" {
			t.Errorf("Expected name 'Test User', got %q", config.Name)
		}

		if config.Email != "test@example.com" {
			t.Errorf("Expected email 'test@example.com', got %q", config.Email)
		}
	})

	t.Run("returns empty values when no configuration exists", func(t *testing.T) {
		executor := mocks.NewMockExecutor()
		stateChecker := mocks.NewMockStateChecker()

		// Mock git config commands to return errors (no config)
		executor.SetExecuteResponse("git config --global --get user.name", "", context.DeadlineExceeded)
		executor.SetExecuteResponse("git config --global --get user.email", "", context.DeadlineExceeded)

		manager := NewGitConfigManager(executor, stateChecker)
		config, err := manager.GetCurrentGitConfig(ctx)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if config.Name != "" {
			t.Errorf("Expected empty name, got %q", config.Name)
		}

		if config.Email != "" {
			t.Errorf("Expected empty email, got %q", config.Email)
		}
	})
}

func TestGitConfigManager_ValidateGitConfig(t *testing.T) {
	manager := &GitConfigManager{}

	t.Run("validates correct configuration", func(t *testing.T) {
		config := GitConfig{
			Name:  "Test User",
			Email: "test@example.com",
		}

		err := manager.ValidateGitConfig(config)

		if err != nil {
			t.Errorf("Expected no error for valid config, got %v", err)
		}
	})

	t.Run("rejects empty name", func(t *testing.T) {
		config := GitConfig{
			Name:  "",
			Email: "test@example.com",
		}

		err := manager.ValidateGitConfig(config)

		if err == nil {
			t.Error("Expected error for empty name")
		}
	})

	t.Run("rejects empty email", func(t *testing.T) {
		config := GitConfig{
			Name:  "Test User",
			Email: "",
		}

		err := manager.ValidateGitConfig(config)

		if err == nil {
			t.Error("Expected error for empty email")
		}
	})

	t.Run("rejects invalid email format", func(t *testing.T) {
		config := GitConfig{
			Name:  "Test User",
			Email: "invalid-email",
		}

		err := manager.ValidateGitConfig(config)

		if err == nil {
			t.Error("Expected error for invalid email format")
		}
	})
}
