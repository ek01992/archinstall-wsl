package system

import (
	"context"
	"testing"

	"archinstall-wsl/internal/system/mocks"
)

func TestFirewallManager_ConfigureFirewall(t *testing.T) {
	ctx := context.Background()

	t.Run("installs and configures ufw when not installed", func(t *testing.T) {
		executor := mocks.NewMockExecutor()
		stateChecker := mocks.NewMockStateChecker()

		// Mock ufw not installed
		stateChecker.SetPackageInstalled("ufw", false)

		// Mock successful commands
		executor.SetDefaultExecuteResponse("", nil)
		executor.SetDefaultExecuteSilentError(nil)

		manager := NewFirewallManager(executor, stateChecker)
		config := DefaultFirewallConfig()

		err := manager.ConfigureFirewall(ctx, config)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("configures ufw when already installed", func(t *testing.T) {
		executor := mocks.NewMockExecutor()
		stateChecker := mocks.NewMockStateChecker()

		// Mock ufw already installed
		stateChecker.SetPackageInstalled("ufw", true)

		// Mock successful commands
		executor.SetDefaultExecuteResponse("", nil)
		executor.SetDefaultExecuteSilentError(nil)

		manager := NewFirewallManager(executor, stateChecker)
		config := DefaultFirewallConfig()

		err := manager.ConfigureFirewall(ctx, config)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("handles ufw installation failure", func(t *testing.T) {
		executor := mocks.NewMockExecutor()
		stateChecker := mocks.NewMockStateChecker()

		// Mock ufw not installed
		stateChecker.SetPackageInstalled("ufw", false)

		// Mock installation failure
		executor.SetExecuteResponse("pacman -S --noconfirm ufw", "", context.DeadlineExceeded)
		executor.SetDefaultExecuteResponse("", nil) // For other commands

		manager := NewFirewallManager(executor, stateChecker)
		config := DefaultFirewallConfig()

		err := manager.ConfigureFirewall(ctx, config)

		if err == nil {
			t.Error("Expected error when ufw installation fails")
		}
	})

	t.Run("handles ufw enable failure", func(t *testing.T) {
		executor := mocks.NewMockExecutor()
		stateChecker := mocks.NewMockStateChecker()

		// Mock ufw already installed
		stateChecker.SetPackageInstalled("ufw", true)

		// Mock successful commands except enable
		executor.SetDefaultExecuteResponse("", nil)
		executor.SetExecuteResponse("ufw --force enable", "", context.DeadlineExceeded)

		manager := NewFirewallManager(executor, stateChecker)
		config := DefaultFirewallConfig()

		err := manager.ConfigureFirewall(ctx, config)

		if err == nil {
			t.Error("Expected error when ufw enable fails")
		}
	})
}

func TestFirewallManager_ValidateFirewallConfig(t *testing.T) {
	manager := &FirewallManager{}

	t.Run("validates correct configuration", func(t *testing.T) {
		config := &FirewallConfig{
			DefaultPolicy: "deny",
			WSLInterface:  "eth0",
		}

		err := manager.ValidateFirewallConfig(config)

		if err != nil {
			t.Errorf("Expected no error for valid config, got %v", err)
		}
	})

	t.Run("rejects nil configuration", func(t *testing.T) {
		err := manager.ValidateFirewallConfig(nil)

		if err == nil {
			t.Error("Expected error for nil config")
		}
	})

	t.Run("rejects invalid default policy", func(t *testing.T) {
		config := &FirewallConfig{
			DefaultPolicy: "invalid",
			WSLInterface:  "eth0",
		}

		err := manager.ValidateFirewallConfig(config)

		if err == nil {
			t.Error("Expected error for invalid default policy")
		}
	})

	t.Run("rejects empty WSL interface", func(t *testing.T) {
		config := &FirewallConfig{
			DefaultPolicy: "deny",
			WSLInterface:  "",
		}

		err := manager.ValidateFirewallConfig(config)

		if err == nil {
			t.Error("Expected error for empty WSL interface")
		}
	})
}

func TestFirewallManager_IsUFWActive(t *testing.T) {
	ctx := context.Background()

	t.Run("detects active ufw", func(t *testing.T) {
		executor := mocks.NewMockExecutor()
		stateChecker := mocks.NewMockStateChecker()

		// Mock active ufw status
		executor.SetExecuteResponse("ufw status", "Status: active", nil)

		manager := NewFirewallManager(executor, stateChecker)
		active, err := manager.IsUFWActive(ctx)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if !active {
			t.Error("Expected ufw to be active")
		}
	})

	t.Run("detects inactive ufw", func(t *testing.T) {
		executor := mocks.NewMockExecutor()
		stateChecker := mocks.NewMockStateChecker()

		// Mock inactive ufw status
		executor.SetExecuteResponse("ufw status", "Status: inactive", nil)

		manager := NewFirewallManager(executor, stateChecker)
		active, err := manager.IsUFWActive(ctx)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if active {
			t.Error("Expected ufw to be inactive")
		}
	})

	t.Run("handles ufw status command failure", func(t *testing.T) {
		executor := mocks.NewMockExecutor()
		stateChecker := mocks.NewMockStateChecker()

		// Mock ufw status command failure
		executor.SetExecuteResponse("ufw status", "", context.DeadlineExceeded)

		manager := NewFirewallManager(executor, stateChecker)
		_, err := manager.IsUFWActive(ctx)

		if err == nil {
			t.Error("Expected error when ufw status command fails")
		}
	})
}

func TestFirewallManager_GetFirewallStatus(t *testing.T) {
	ctx := context.Background()

	t.Run("retrieves firewall status", func(t *testing.T) {
		executor := mocks.NewMockExecutor()
		stateChecker := mocks.NewMockStateChecker()

		// Mock ufw status output
		expectedStatus := "Status: active\nDefault: deny (incoming), allow (outgoing)"
		executor.SetExecuteResponse("ufw status verbose", expectedStatus, nil)

		manager := NewFirewallManager(executor, stateChecker)
		status, err := manager.GetFirewallStatus(ctx)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if status != expectedStatus {
			t.Errorf("Expected status %q, got %q", expectedStatus, status)
		}
	})

	t.Run("handles status command failure", func(t *testing.T) {
		executor := mocks.NewMockExecutor()
		stateChecker := mocks.NewMockStateChecker()

		// Mock ufw status command failure
		executor.SetExecuteResponse("ufw status verbose", "", context.DeadlineExceeded)

		manager := NewFirewallManager(executor, stateChecker)
		_, err := manager.GetFirewallStatus(ctx)

		if err == nil {
			t.Error("Expected error when ufw status command fails")
		}
	})
}

func TestDefaultFirewallConfig(t *testing.T) {
	config := DefaultFirewallConfig()

	if config.DefaultPolicy != "deny" {
		t.Errorf("Expected default policy 'deny', got %s", config.DefaultPolicy)
	}

	if config.WSLInterface != "eth0" {
		t.Errorf("Expected WSL interface 'eth0', got %s", config.WSLInterface)
	}
}
