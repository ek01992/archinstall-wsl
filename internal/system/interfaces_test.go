package system

import (
	"context"
	"testing"
)

// Test interfaces can be implemented
func TestExecutorInterface(t *testing.T) {
	var _ Executor = (*MockExecutor)(nil)
}

func TestStateCheckerInterface(t *testing.T) {
	var _ StateChecker = (*MockStateChecker)(nil)
}

// Mock implementations for testing interface compliance
type MockExecutor struct{}

func (m *MockExecutor) Execute(ctx context.Context, command string) (string, error) {
	return "", nil
}

func (m *MockExecutor) ExecuteWithOutput(ctx context.Context, command string) (stdout string, stderr string, err error) {
	return "", "", nil
}

func (m *MockExecutor) ExecuteSilent(ctx context.Context, command string) error {
	return nil
}

type MockStateChecker struct{}

func (m *MockStateChecker) UserExists(username string) (bool, error) {
	return false, nil
}

func (m *MockStateChecker) PackageIsInstalled(pkgName string) (bool, error) {
	return false, nil
}

func (m *MockStateChecker) FileContains(path, content string) (bool, error) {
	return false, nil
}

func (m *MockStateChecker) IsSudoer(username string) (bool, error) {
	return false, nil
}

func (m *MockStateChecker) DirectoryExists(path string) (bool, error) {
	return false, nil
}

func (m *MockStateChecker) FileExists(path string) (bool, error) {
	return false, nil
}

func (m *MockStateChecker) ServiceIsRunning(serviceName string) (bool, error) {
	return false, nil
}

func (m *MockStateChecker) ServiceIsEnabled(serviceName string) (bool, error) {
	return false, nil
}
