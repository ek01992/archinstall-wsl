package system

import (
	"context"
	"testing"

	"archinstall-wsl/internal/system/mocks"
)

func TestSSHManager_IntegrateSSHKeys(t *testing.T) {
	ctx := context.Background()

	t.Run("creates SSH directory and copies keys when source exists", func(t *testing.T) {
		executor := mocks.NewMockExecutor()
		stateChecker := mocks.NewMockStateChecker()

		// Mock Windows username
		executor.SetExecuteResponse("cmd.exe /c echo %USERNAME%", "testuser", nil)

		// Mock source directory exists
		stateChecker.SetDirectoryExists("/mnt/c/Users/testuser/.ssh", true)

		// Mock target directory doesn't exist
		stateChecker.SetDirectoryExists("/home/testuser/.ssh", false)

		// Mock file listing
		executor.SetExecuteResponse("ls -la /mnt/c/Users/testuser/.ssh",
			"total 8\ndrwx------ 2 testuser testuser 4096 Jan 1 12:00 .\n"+
				"drwx------ 2 testuser testuser 4096 Jan 1 12:00 ..\n"+
				"-rw------- 1 testuser testuser 1679 Jan 1 12:00 id_rsa\n"+
				"-rw-r--r-- 1 testuser testuser  393 Jan 1 12:00 id_rsa.pub\n", nil)

		// Mock target files don't exist
		stateChecker.SetFileExists("/home/testuser/.ssh/id_rsa", false)
		stateChecker.SetFileExists("/home/testuser/.ssh/id_rsa.pub", false)

		// Mock successful commands
		executor.SetDefaultExecuteResponse("", nil)
		executor.SetDefaultExecuteSilentError(nil)

		manager := NewSSHManager(executor, stateChecker)
		result, err := manager.IntegrateSSHKeys(ctx, "testuser")

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if !result.KeysCopied {
			t.Error("Expected keys to be copied")
		}

		if len(result.FilesCopied) != 2 {
			t.Errorf("Expected 2 files copied, got %d", len(result.FilesCopied))
		}

		expectedFiles := []string{"id_rsa", "id_rsa.pub"}
		for _, expected := range expectedFiles {
			found := false
			for _, copied := range result.FilesCopied {
				if copied == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected file %s to be copied", expected)
			}
		}
	})

	t.Run("skips existing files", func(t *testing.T) {
		executor := mocks.NewMockExecutor()
		stateChecker := mocks.NewMockStateChecker()

		// Mock Windows username
		executor.SetExecuteResponse("cmd.exe /c echo %USERNAME%", "testuser", nil)

		// Mock source and target directories exist
		stateChecker.SetDirectoryExists("/mnt/c/Users/testuser/.ssh", true)
		stateChecker.SetDirectoryExists("/home/testuser/.ssh", true)

		// Mock file listing
		executor.SetExecuteResponse("ls -la /mnt/c/Users/testuser/.ssh",
			"total 8\ndrwx------ 2 testuser testuser 4096 Jan 1 12:00 .\n"+
				"drwx------ 2 testuser testuser 4096 Jan 1 12:00 ..\n"+
				"-rw------- 1 testuser testuser 1679 Jan 1 12:00 id_rsa\n", nil)

		// Mock target file already exists
		stateChecker.SetFileExists("/home/testuser/.ssh/id_rsa", true)

		manager := NewSSHManager(executor, stateChecker)
		result, err := manager.IntegrateSSHKeys(ctx, "testuser")

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if result.KeysCopied {
			t.Error("Expected no keys to be copied")
		}

		if len(result.FilesCopied) != 0 {
			t.Errorf("Expected 0 files copied, got %d", len(result.FilesCopied))
		}
	})

	t.Run("handles missing source directory", func(t *testing.T) {
		executor := mocks.NewMockExecutor()
		stateChecker := mocks.NewMockStateChecker()

		// Mock Windows username
		executor.SetExecuteResponse("cmd.exe /c echo %USERNAME%", "testuser", nil)

		// Mock source directory doesn't exist
		stateChecker.SetDirectoryExists("/mnt/c/Users/testuser/.ssh", false)

		manager := NewSSHManager(executor, stateChecker)
		result, err := manager.IntegrateSSHKeys(ctx, "testuser")

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if result.KeysCopied {
			t.Error("Expected no keys to be copied")
		}

		if len(result.Errors) == 0 {
			t.Error("Expected error message about missing source directory")
		}
	})

	t.Run("handles Windows username detection failure", func(t *testing.T) {
		executor := mocks.NewMockExecutor()
		stateChecker := mocks.NewMockStateChecker()

		// Mock Windows username command failure
		executor.SetExecuteResponse("cmd.exe /c echo %USERNAME%", "", context.DeadlineExceeded)

		// Mock source directory exists with fallback path
		stateChecker.SetDirectoryExists("/mnt/c/Users/Administrator/.ssh", true)

		manager := NewSSHManager(executor, stateChecker)
		result, err := manager.IntegrateSSHKeys(ctx, "testuser")

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Should use fallback path
		if result.SourcePath != "/mnt/c/Users/Administrator/.ssh" {
			t.Errorf("Expected fallback source path, got %s", result.SourcePath)
		}
	})
}

func TestSSHManager_ValidateSSHSetup(t *testing.T) {
	ctx := context.Background()

	t.Run("validates correct SSH setup", func(t *testing.T) {
		executor := mocks.NewMockExecutor()
		stateChecker := mocks.NewMockStateChecker()

		// Mock SSH directory exists
		stateChecker.SetDirectoryExists("/home/testuser/.ssh", true)

		// Mock correct permissions
		executor.SetExecuteResponse("stat -c %a /home/testuser/.ssh", "700", nil)

		manager := NewSSHManager(executor, stateChecker)
		err := manager.ValidateSSHSetup(ctx, "testuser")

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("fails when SSH directory doesn't exist", func(t *testing.T) {
		executor := mocks.NewMockExecutor()
		stateChecker := mocks.NewMockStateChecker()

		// Mock SSH directory doesn't exist
		stateChecker.SetDirectoryExists("/home/testuser/.ssh", false)

		manager := NewSSHManager(executor, stateChecker)
		err := manager.ValidateSSHSetup(ctx, "testuser")

		if err == nil {
			t.Error("Expected error when SSH directory doesn't exist")
		}
	})

	t.Run("fails when SSH directory has wrong permissions", func(t *testing.T) {
		executor := mocks.NewMockExecutor()
		stateChecker := mocks.NewMockStateChecker()

		// Mock SSH directory exists
		stateChecker.SetDirectoryExists("/home/testuser/.ssh", true)

		// Mock wrong permissions
		executor.SetExecuteResponse("stat -c %a /home/testuser/.ssh", "755", nil)

		manager := NewSSHManager(executor, stateChecker)
		err := manager.ValidateSSHSetup(ctx, "testuser")

		if err == nil {
			t.Error("Expected error when SSH directory has wrong permissions")
		}
	})
}
