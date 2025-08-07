package system

import (
	"context"
	"fmt"
	"strings"
)

// FirewallManager handles firewall configuration using ufw
type FirewallManager struct {
	executor     Executor
	stateChecker StateChecker
}

// NewFirewallManager creates a new FirewallManager
func NewFirewallManager(executor Executor, stateChecker StateChecker) *FirewallManager {
	return &FirewallManager{
		executor:     executor,
		stateChecker: stateChecker,
	}
}

// FirewallConfig represents firewall configuration
type FirewallConfig struct {
	DefaultPolicy string // "deny" or "allow"
	WSLInterface  string // WSL network interface name
}

// DefaultFirewallConfig returns the default firewall configuration
func DefaultFirewallConfig() *FirewallConfig {
	return &FirewallConfig{
		DefaultPolicy: "deny",
		WSLInterface:  "eth0", // Default WSL interface
	}
}

// ConfigureFirewall installs and configures ufw firewall
func (m *FirewallManager) ConfigureFirewall(ctx context.Context, config *FirewallConfig) error {
	// Check if ufw is installed
	installed, err := m.stateChecker.PackageIsInstalled("ufw")
	if err != nil {
		return fmt.Errorf("failed to check if ufw is installed: %w", err)
	}

	// Install ufw if not installed
	if !installed {
		err = m.installUFW(ctx)
		if err != nil {
			return fmt.Errorf("failed to install ufw: %w", err)
		}
	}

	// Reset ufw to default state
	err = m.resetUFW(ctx)
	if err != nil {
		return fmt.Errorf("failed to reset ufw: %w", err)
	}

	// Configure default policy
	err = m.setDefaultPolicy(ctx, config.DefaultPolicy)
	if err != nil {
		return fmt.Errorf("failed to set default policy: %w", err)
	}

	// Allow SSH
	err = m.allowSSH(ctx)
	if err != nil {
		return fmt.Errorf("failed to allow SSH: %w", err)
	}

	// Allow WSL host-guest communication
	err = m.allowWSLCommunication(ctx, config.WSLInterface)
	if err != nil {
		return fmt.Errorf("failed to allow WSL communication: %w", err)
	}

	// Enable ufw
	err = m.enableUFW(ctx)
	if err != nil {
		return fmt.Errorf("failed to enable ufw: %w", err)
	}

	return nil
}

// installUFW installs the ufw package
func (m *FirewallManager) installUFW(ctx context.Context) error {
	// Update package list
	updateCmd := "pacman -Sy --noconfirm"
	_, err := m.executor.Execute(ctx, updateCmd)
	if err != nil {
		return fmt.Errorf("failed to update package list: %w", err)
	}

	// Install ufw
	installCmd := "pacman -S --noconfirm ufw"
	_, err = m.executor.Execute(ctx, installCmd)
	if err != nil {
		return fmt.Errorf("failed to install ufw: %w", err)
	}

	return nil
}

// resetUFW resets ufw to default state
func (m *FirewallManager) resetUFW(ctx context.Context) error {
	// Disable ufw first
	disableCmd := "ufw --force disable"
	_, err := m.executor.Execute(ctx, disableCmd)
	if err != nil {
		return fmt.Errorf("failed to disable ufw: %w", err)
	}

	// Reset ufw
	resetCmd := "ufw --force reset"
	_, err = m.executor.Execute(ctx, resetCmd)
	if err != nil {
		return fmt.Errorf("failed to reset ufw: %w", err)
	}

	return nil
}

// setDefaultPolicy sets the default ufw policy
func (m *FirewallManager) setDefaultPolicy(ctx context.Context, policy string) error {
	// Set default incoming policy
	// Corrected line:
	incomingCmd := "ufw default deny incoming"
	if policy == "allow" {
		incomingCmd = "ufw default allow incoming"
	}

	_, err := m.executor.Execute(ctx, incomingCmd)
	if err != nil {
		return fmt.Errorf("failed to set default incoming policy: %w", err)
	}

	// Set default outgoing policy (always allow)
	outgoingCmd := "ufw default allow outgoing"
	_, err = m.executor.Execute(ctx, outgoingCmd)
	if err != nil {
		return fmt.Errorf("failed to set default outgoing policy: %w", err)
	}

	return nil
}

// allowSSH allows SSH connections
func (m *FirewallManager) allowSSH(ctx context.Context) error {
	sshCmd := "ufw allow ssh"
	_, err := m.executor.Execute(ctx, sshCmd)
	if err != nil {
		return fmt.Errorf("failed to allow SSH: %w", err)
	}

	return nil
}

// allowWSLCommunication allows communication on WSL interface
func (m *FirewallManager) allowWSLCommunication(ctx context.Context, interfaceName string) error {
	// Allow all traffic on WSL interface
	interfaceCmd := fmt.Sprintf("ufw allow in on %s", interfaceName)
	_, err := m.executor.Execute(ctx, interfaceCmd)
	if err != nil {
		return fmt.Errorf("failed to allow traffic on WSL interface: %w", err)
	}

	return nil
}

// enableUFW enables ufw firewall
func (m *FirewallManager) enableUFW(ctx context.Context) error {
	enableCmd := "ufw --force enable"
	_, err := m.executor.Execute(ctx, enableCmd)
	if err != nil {
		return fmt.Errorf("failed to enable ufw: %w", err)
	}

	return nil
}

// GetFirewallStatus gets the current firewall status
func (m *FirewallManager) GetFirewallStatus(ctx context.Context) (string, error) {
	statusCmd := "ufw status verbose"
	output, err := m.executor.Execute(ctx, statusCmd)
	if err != nil {
		return "", fmt.Errorf("failed to get ufw status: %w", err)
	}

	return output, nil
}

// ValidateFirewallConfig validates firewall configuration
func (m *FirewallManager) ValidateFirewallConfig(config *FirewallConfig) error {
	if config == nil {
		return fmt.Errorf("firewall config cannot be nil")
	}

	if config.DefaultPolicy != "deny" && config.DefaultPolicy != "allow" {
		return fmt.Errorf("default policy must be 'deny' or 'allow', got %s", config.DefaultPolicy)
	}

	if config.WSLInterface == "" {
		return fmt.Errorf("WSL interface cannot be empty")
	}

	return nil
}

// IsUFWActive checks if ufw is active
func (m *FirewallManager) IsUFWActive(ctx context.Context) (bool, error) {
	statusCmd := "ufw status"
	output, err := m.executor.Execute(ctx, statusCmd)
	if err != nil {
		return false, fmt.Errorf("failed to check ufw status: %w", err)
	}
	// Check for the exact phrase "status: active" to avoid matching "inactive"
	return strings.Contains(strings.ToLower(output), "status: active"), nil
}
