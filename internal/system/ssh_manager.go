package system

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
)

// SSHManager handles SSH key integration between Windows host and WSL
type SSHManager struct {
	executor     Executor
	stateChecker StateChecker
}

// NewSSHManager creates a new SSHManager
func NewSSHManager(executor Executor, stateChecker StateChecker) *SSHManager {
	return &SSHManager{
		executor:     executor,
		stateChecker: stateChecker,
	}
}

// SSHKeyResult represents the result of SSH key integration
type SSHKeyResult struct {
	KeysCopied  bool
	SourcePath  string
	TargetPath  string
	FilesCopied []string
	Errors      []string
}

// IntegrateSSHKeys copies SSH keys from Windows host to WSL
func (m *SSHManager) IntegrateSSHKeys(ctx context.Context, username string) (*SSHKeyResult, error) {
	result := &SSHKeyResult{
		FilesCopied: []string{},
		Errors:      []string{},
	}

	// Determine source and target paths
	sourcePath := m.getWindowsSSHPath()
	targetPath := fmt.Sprintf("/home/%s/.ssh", username)

	result.SourcePath = sourcePath
	result.TargetPath = targetPath

	// Check if source directory exists
	sourceExists, err := m.stateChecker.DirectoryExists(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("failed to check source SSH directory: %w", err)
	}

	if !sourceExists {
		result.Errors = append(result.Errors, fmt.Sprintf("Source SSH directory does not exist: %s", sourcePath))
		return result, nil
	}

	// Check if target directory exists
	targetExists, err := m.stateChecker.DirectoryExists(targetPath)
	if err != nil {
		return nil, fmt.Errorf("failed to check target SSH directory: %w", err)
	}

	// Create target directory if it doesn't exist
	if !targetExists {
		err = m.createSSHDirectory(ctx, targetPath, username)
		if err != nil {
			return nil, fmt.Errorf("failed to create SSH directory: %w", err)
		}
		result.KeysCopied = true
	}

	// Copy SSH files
	err = m.copySSHFiles(ctx, sourcePath, targetPath, username, result)
	if err != nil {
		return nil, fmt.Errorf("failed to copy SSH files: %w", err)
	}

	if len(result.FilesCopied) > 0 {
		result.KeysCopied = true
	}

	return result, nil
}

// getWindowsSSHPath returns the Windows SSH directory path
func (m *SSHManager) getWindowsSSHPath() string {
	// In WSL, Windows drives are mounted under /mnt/
	// We need to get the Windows username to construct the path
	cmd := "cmd.exe /c echo %USERNAME%"
	output, err := m.executor.Execute(context.Background(), cmd)
	if err != nil {
		// Fallback to a common path
		return "/mnt/c/Users/Administrator/.ssh"
	}

	windowsUsername := strings.TrimSpace(output)
	if windowsUsername == "" {
		windowsUsername = "Administrator"
	}

	return fmt.Sprintf("/mnt/c/Users/%s/.ssh", windowsUsername)
}

// createSSHDirectory creates the SSH directory with proper permissions
func (m *SSHManager) createSSHDirectory(ctx context.Context, path, username string) error {
	// Create directory
	createCmd := fmt.Sprintf("mkdir -p %s", path)
	_, err := m.executor.Execute(ctx, createCmd)
	if err != nil {
		return fmt.Errorf("failed to create SSH directory: %w", err)
	}

	// Set ownership
	chownCmd := fmt.Sprintf("chown %s:%s %s", username, username, path)
	_, err = m.executor.Execute(ctx, chownCmd)
	if err != nil {
		return fmt.Errorf("failed to set SSH directory ownership: %w", err)
	}

	// Set permissions (700)
	chmodCmd := fmt.Sprintf("chmod 700 %s", path)
	_, err = m.executor.Execute(ctx, chmodCmd)
	if err != nil {
		return fmt.Errorf("failed to set SSH directory permissions: %w", err)
	}

	return nil
}

// copySSHFiles copies SSH files from Windows to WSL
func (m *SSHManager) copySSHFiles(ctx context.Context, sourcePath, targetPath, username string, result *SSHKeyResult) error {
	// List files in source directory
	listCmd := fmt.Sprintf("ls -la %s", sourcePath)
	output, err := m.executor.Execute(ctx, listCmd)
	if err != nil {
		return fmt.Errorf("failed to list SSH files: %w", err)
	}

	// Parse file list and copy each file
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Extract filename from ls output
		fields := strings.Fields(line)
		if len(fields) < 9 {
			continue
		}

		filename := fields[8]
		if filename == "." || filename == ".." {
			continue
		}

		sourceFile := filepath.Join(sourcePath, filename)
		targetFile := filepath.Join(targetPath, filename)

		// Check if target file already exists
		targetExists, err := m.stateChecker.FileExists(targetFile)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to check target file %s: %v", targetFile, err))
			continue
		}

		if targetExists {
			// Skip if file already exists
			continue
		}

		// Copy file
		copyCmd := fmt.Sprintf("cp %s %s", sourceFile, targetFile)
		_, err = m.executor.Execute(ctx, copyCmd)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to copy %s: %v", filename, err))
			continue
		}

		// Set ownership
		chownCmd := fmt.Sprintf("chown %s:%s %s", username, username, targetFile)
		_, err = m.executor.Execute(ctx, chownCmd)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to set ownership for %s: %v", filename, err))
			continue
		}

		// Set permissions based on file type
		var chmodCmd string
		if strings.HasSuffix(filename, ".pub") {
			// Public keys get 644 permissions
			chmodCmd = fmt.Sprintf("chmod 644 %s", targetFile)
		} else {
			// Private keys get 600 permissions
			chmodCmd = fmt.Sprintf("chmod 600 %s", targetFile)
		}

		_, err = m.executor.Execute(ctx, chmodCmd)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to set permissions for %s: %v", filename, err))
			continue
		}

		result.FilesCopied = append(result.FilesCopied, filename)
	}

	return nil
}

// ValidateSSHSetup validates the SSH setup
func (m *SSHManager) ValidateSSHSetup(ctx context.Context, username string) error {
	sshDir := fmt.Sprintf("/home/%s/.ssh", username)

	// Check if SSH directory exists
	exists, err := m.stateChecker.DirectoryExists(sshDir)
	if err != nil {
		return fmt.Errorf("failed to check SSH directory: %w", err)
	}

	if !exists {
		return fmt.Errorf("SSH directory does not exist: %s", sshDir)
	}

	// Check SSH directory permissions
	statCmd := fmt.Sprintf("stat -c %%a %s", sshDir)
	output, err := m.executor.Execute(ctx, statCmd)
	if err != nil {
		return fmt.Errorf("failed to check SSH directory permissions: %w", err)
	}

	permissions := strings.TrimSpace(output)
	if permissions != "700" {
		return fmt.Errorf("SSH directory has incorrect permissions: %s (expected 700)", permissions)
	}

	return nil
}
