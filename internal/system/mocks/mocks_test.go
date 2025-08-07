package mocks

import (
	"context"
	"errors"
	"testing"
)

func TestMockExecutor(t *testing.T) {
	executor := NewMockExecutor()
	ctx := context.Background()

	t.Run("Execute returns configured response", func(t *testing.T) {
		expectedOutput := "test output"
		executor.SetExecuteResponse("test command", expectedOutput, nil)

		output, err := executor.Execute(ctx, "test command")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if output != expectedOutput {
			t.Errorf("Expected output %q, got %q", expectedOutput, output)
		}
	})

	t.Run("Execute returns default response when no specific response set", func(t *testing.T) {
		expectedOutput := "default output"
		executor.SetDefaultExecuteResponse(expectedOutput, nil)

		output, err := executor.Execute(ctx, "unknown command")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if output != expectedOutput {
			t.Errorf("Expected output %q, got %q", expectedOutput, output)
		}
	})

	t.Run("ExecuteSilent returns configured error", func(t *testing.T) {
		expectedErr := errors.New("test error")
		executor.SetExecuteSilentError("test command", expectedErr)

		err := executor.ExecuteSilent(ctx, "test command")
		if err != expectedErr {
			t.Errorf("Expected error %v, got %v", expectedErr, err)
		}
	})

	t.Run("ExecuteSilent returns default error when no specific error set", func(t *testing.T) {
		expectedErr := errors.New("default error")
		executor.SetDefaultExecuteSilentError(expectedErr)

		err := executor.ExecuteSilent(ctx, "unknown command")
		if err != expectedErr {
			t.Errorf("Expected error %v, got %v", expectedErr, err)
		}
	})
}

func TestMockStateChecker(t *testing.T) {
	checker := NewMockStateChecker()

	t.Run("UserExists returns configured value", func(t *testing.T) {
		checker.SetUserExists("testuser", true)

		exists, err := checker.UserExists("testuser")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if !exists {
			t.Error("Expected user to exist")
		}
	})

	t.Run("UserExists returns default value when no specific value set", func(t *testing.T) {
		checker.SetDefaultUserExists(false)

		exists, err := checker.UserExists("unknownuser")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if exists {
			t.Error("Expected user to not exist")
		}
	})

	t.Run("PackageIsInstalled returns configured value", func(t *testing.T) {
		checker.SetPackageInstalled("testpkg", true)

		installed, err := checker.PackageIsInstalled("testpkg")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if !installed {
			t.Error("Expected package to be installed")
		}
	})

	t.Run("FileContains returns configured value", func(t *testing.T) {
		checker.SetFileContains("/test/file", "test content", true)

		contains, err := checker.FileContains("/test/file", "test content")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if !contains {
			t.Error("Expected file to contain content")
		}
	})

	t.Run("IsSudoer returns configured value", func(t *testing.T) {
		checker.SetIsSudoer("testuser", true)

		isSudoer, err := checker.IsSudoer("testuser")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if !isSudoer {
			t.Error("Expected user to be sudoer")
		}
	})

	t.Run("DirectoryExists returns configured value", func(t *testing.T) {
		checker.SetDirectoryExists("/test/dir", true)

		exists, err := checker.DirectoryExists("/test/dir")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if !exists {
			t.Error("Expected directory to exist")
		}
	})

	t.Run("FileExists returns configured value", func(t *testing.T) {
		checker.SetFileExists("/test/file", true)

		exists, err := checker.FileExists("/test/file")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if !exists {
			t.Error("Expected file to exist")
		}
	})

	t.Run("ServiceIsRunning returns configured value", func(t *testing.T) {
		checker.SetServiceRunning("testservice", true)

		running, err := checker.ServiceIsRunning("testservice")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if !running {
			t.Error("Expected service to be running")
		}
	})

	t.Run("ServiceIsEnabled returns configured value", func(t *testing.T) {
		checker.SetServiceEnabled("testservice", true)

		enabled, err := checker.ServiceIsEnabled("testservice")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if !enabled {
			t.Error("Expected service to be enabled")
		}
	})
}
