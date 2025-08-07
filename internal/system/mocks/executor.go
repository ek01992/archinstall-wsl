package mocks

import (
	"context"
	"sync"
)

// MockExecutor is a configurable mock implementation of the Executor interface
type MockExecutor struct {
	mu sync.RWMutex

	// Configurable responses
	executeResponses    map[string]executeResponse
	executeSilentErrors map[string]error

	// Default responses
	defaultExecuteResponse    executeResponse
	defaultExecuteSilentError error
}

type executeResponse struct {
	output string
	err    error
}

// NewMockExecutor creates a new MockExecutor
func NewMockExecutor() *MockExecutor {
	return &MockExecutor{
		executeResponses:    make(map[string]executeResponse),
		executeSilentErrors: make(map[string]error),
	}
}

// SetExecuteResponse sets the response for a specific command
func (m *MockExecutor) SetExecuteResponse(command string, output string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.executeResponses[command] = executeResponse{output: output, err: err}
}

// SetDefaultExecuteResponse sets the default response for Execute calls
func (m *MockExecutor) SetDefaultExecuteResponse(output string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.defaultExecuteResponse = executeResponse{output: output, err: err}
}

// SetExecuteSilentError sets the error for a specific command in ExecuteSilent
func (m *MockExecutor) SetExecuteSilentError(command string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.executeSilentErrors[command] = err
}

// SetDefaultExecuteSilentError sets the default error for ExecuteSilent calls
func (m *MockExecutor) SetDefaultExecuteSilentError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.defaultExecuteSilentError = err
}

// Execute implements the Executor interface
func (m *MockExecutor) Execute(ctx context.Context, command string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if response, exists := m.executeResponses[command]; exists {
		return response.output, response.err
	}

	return m.defaultExecuteResponse.output, m.defaultExecuteResponse.err
}

// ExecuteWithOutput implements the Executor interface
func (m *MockExecutor) ExecuteWithOutput(ctx context.Context, command string) (stdout string, stderr string, err error) {
	output, err := m.Execute(ctx, command)
	if err != nil {
		return "", "", err
	}
	return output, "", nil
}

// ExecuteSilent implements the Executor interface
func (m *MockExecutor) ExecuteSilent(ctx context.Context, command string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if err, exists := m.executeSilentErrors[command]; exists {
		return err
	}

	return m.defaultExecuteSilentError
}
