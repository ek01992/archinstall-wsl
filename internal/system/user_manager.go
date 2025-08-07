package system

import (
	"context"
	"fmt"
	"strings"
)

// UserManager handles user creation and management operations
type UserManager struct {
	executor     Executor
	stateChecker StateChecker
}

// NewUserManager creates a new UserManager
func NewUserManager(executor Executor, stateChecker StateChecker) *UserManager {
	return &UserManager{
		executor:     executor,
		stateChecker: stateChecker,
	}
}

// UserCreationResult represents the result of a user creation operation
type UserCreationResult struct {
	UserCreated  bool
	Commands     []string
	UndoCommands []string
}

// CreateUserWithSudo creates a user with sudo privileges in a transactional manner
func (m *UserManager) CreateUserWithSudo(ctx context.Context, username string) (*UserCreationResult, error) {
	result := &UserCreationResult{
		Commands:     []string{},
		UndoCommands: []string{},
	}

	// Check if user already exists
	exists, err := m.stateChecker.UserExists(username)
	if err != nil {
		return nil, fmt.Errorf("failed to check if user exists: %w", err)
	}

	if exists {
		// User already exists, check if they have sudo privileges
		isSudoer, err := m.stateChecker.IsSudoer(username)
		if err != nil {
			return nil, fmt.Errorf("failed to check sudo privileges: %w", err)
		}

		if isSudoer {
			// User exists and has sudo privileges, nothing to do
			return result, nil
		}

		// User exists but doesn't have sudo privileges, add them
		return m.addSudoPrivileges(ctx, username)
	}

	// User doesn't exist, create them with sudo privileges
	return m.createNewUserWithSudo(ctx, username)
}

// createNewUserWithSudo creates a new user with sudo privileges
func (m *UserManager) createNewUserWithSudo(ctx context.Context, username string) (*UserCreationResult, error) {
	result := &UserCreationResult{
		UserCreated:  true,
		Commands:     []string{},
		UndoCommands: []string{},
	}

	// Create user
	createUserCmd := fmt.Sprintf("useradd -m -s /bin/bash %s", username)
	_, err := m.executor.Execute(ctx, createUserCmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	result.Commands = append(result.Commands, createUserCmd)
	result.UndoCommands = append([]string{fmt.Sprintf("userdel -r %s", username)}, result.UndoCommands...)

	// Add user to wheel group
	addToWheelCmd := fmt.Sprintf("usermod -aG wheel %s", username)
	_, err = m.executor.Execute(ctx, addToWheelCmd)
	if err != nil {
		// Rollback user creation
		m.rollbackUserCreation(ctx, username)
		return nil, fmt.Errorf("failed to add user to wheel group: %w", err)
	}
	result.Commands = append(result.Commands, addToWheelCmd)

	// Configure passwordless sudo for wheel group
	sudoConfig := `%wheel ALL=(ALL) NOPASSWD: ALL`
	sudoConfigCmd := fmt.Sprintf("echo '%s' > /etc/sudoers.d/wheel", sudoConfig)
	_, err = m.executor.Execute(ctx, sudoConfigCmd)
	if err != nil {
		// Rollback user creation and wheel group
		m.rollbackUserCreation(ctx, username)
		return nil, fmt.Errorf("failed to configure sudo: %w", err)
	}
	result.Commands = append(result.Commands, sudoConfigCmd)
	result.UndoCommands = append([]string{"rm -f /etc/sudoers.d/wheel"}, result.UndoCommands...)

	return result, nil
}

// addSudoPrivileges adds sudo privileges to an existing user
func (m *UserManager) addSudoPrivileges(ctx context.Context, username string) (*UserCreationResult, error) {
	result := &UserCreationResult{
		UserCreated:  false,
		Commands:     []string{},
		UndoCommands: []string{},
	}

	// Check if user is in wheel group
	checkWheelCmd := fmt.Sprintf("groups %s", username)
	output, err := m.executor.Execute(ctx, checkWheelCmd)
	if err != nil {
		return nil, fmt.Errorf("failed to check user groups: %w", err)
	}

	if !strings.Contains(output, "wheel") {
		// Add user to wheel group
		addToWheelCmd := fmt.Sprintf("usermod -aG wheel %s", username)
		_, err = m.executor.Execute(ctx, addToWheelCmd)
		if err != nil {
			return nil, fmt.Errorf("failed to add user to wheel group: %w", err)
		}
		result.Commands = append(result.Commands, addToWheelCmd)
	}

	// Check if sudoers.d/wheel exists
	sudoersExists, err := m.stateChecker.FileExists("/etc/sudoers.d/wheel")
	if err != nil {
		return nil, fmt.Errorf("failed to check sudoers configuration: %w", err)
	}

	if !sudoersExists {
		// Configure passwordless sudo for wheel group
		sudoConfig := `%wheel ALL=(ALL) NOPASSWD: ALL`
		sudoConfigCmd := fmt.Sprintf("echo '%s' > /etc/sudoers.d/wheel", sudoConfig)
		_, err = m.executor.Execute(ctx, sudoConfigCmd)
		if err != nil {
			return nil, fmt.Errorf("failed to configure sudo: %w", err)
		}
		result.Commands = append(result.Commands, sudoConfigCmd)
		result.UndoCommands = append(result.UndoCommands, "rm -f /etc/sudoers.d/wheel")
	}

	return result, nil
}

// rollbackUserCreation rolls back user creation in case of failure
func (m *UserManager) rollbackUserCreation(ctx context.Context, username string) {
	// Remove user and home directory
	rollbackCmd := fmt.Sprintf("userdel -r %s", username)
	m.executor.ExecuteSilent(ctx, rollbackCmd)
}

// UndoUserCreation executes the undo commands to revert user creation
func (m *UserManager) UndoUserCreation(ctx context.Context, result *UserCreationResult) error {
	for _, cmd := range result.UndoCommands {
		err := m.executor.ExecuteSilent(ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to execute undo command '%s': %w", cmd, err)
		}
	}
	return nil
}
