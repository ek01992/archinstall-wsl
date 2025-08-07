package system

import (
	"context"
	"fmt"
	"strings"
)

// GitConfigManager handles Git configuration operations
type GitConfigManager struct {
	executor     Executor
	stateChecker StateChecker
}

// NewGitConfigManager creates a new GitConfigManager
func NewGitConfigManager(executor Executor, stateChecker StateChecker) *GitConfigManager {
	return &GitConfigManager{
		executor:     executor,
		stateChecker: stateChecker,
	}
}

// GitConfig represents Git user configuration
type GitConfig struct {
	Name  string
	Email string
}

// ConfigureGitUser configures Git user name and email globally
func (m *GitConfigManager) ConfigureGitUser(ctx context.Context, config GitConfig) error {
	// Check if git config already exists
	existingName, err := m.getGitConfig(ctx, "user.name")
	if err != nil {
		return fmt.Errorf("failed to check existing git user.name: %w", err)
	}

	existingEmail, err := m.getGitConfig(ctx, "user.email")
	if err != nil {
		return fmt.Errorf("failed to check existing git user.email: %w", err)
	}

	// Configure user.name if not set or different
	if existingName == "" || existingName != config.Name {
		err = m.setGitConfig(ctx, "user.name", config.Name)
		if err != nil {
			return fmt.Errorf("failed to set git user.name: %w", err)
		}
	}

	// Configure user.email if not set or different
	if existingEmail == "" || existingEmail != config.Email {
		err = m.setGitConfig(ctx, "user.email", config.Email)
		if err != nil {
			return fmt.Errorf("failed to set git user.email: %w", err)
		}
	}

	return nil
}

// GetCurrentGitConfig retrieves the current Git configuration
func (m *GitConfigManager) GetCurrentGitConfig(ctx context.Context) (*GitConfig, error) {
	name, err := m.getGitConfig(ctx, "user.name")
	if err != nil {
		return nil, fmt.Errorf("failed to get git user.name: %w", err)
	}

	email, err := m.getGitConfig(ctx, "user.email")
	if err != nil {
		return nil, fmt.Errorf("failed to get git user.email: %w", err)
	}

	return &GitConfig{
		Name:  name,
		Email: email,
	}, nil
}

// getGitConfig retrieves a Git configuration value
func (m *GitConfigManager) getGitConfig(ctx context.Context, key string) (string, error) {
	cmd := fmt.Sprintf("git config --global --get %s", key)
	output, err := m.executor.Execute(ctx, cmd)
	if err != nil {
		// Git config --get returns error if key doesn't exist, which is expected
		return "", nil
	}
	return strings.TrimSpace(output), nil
}

// setGitConfig sets a Git configuration value
func (m *GitConfigManager) setGitConfig(ctx context.Context, key, value string) error {
	cmd := fmt.Sprintf("git config --global %s %q", key, value)
	_, err := m.executor.Execute(ctx, cmd)
	return err
}

// ValidateGitConfig validates Git configuration values
func (m *GitConfigManager) ValidateGitConfig(config GitConfig) error {
	if config.Name == "" {
		return fmt.Errorf("git user.name cannot be empty")
	}

	if config.Email == "" {
		return fmt.Errorf("git user.email cannot be empty")
	}

	// Basic email validation
	if !strings.Contains(config.Email, "@") {
		return fmt.Errorf("git user.email must be a valid email address")
	}

	return nil
}
