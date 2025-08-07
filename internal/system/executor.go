package system

import (
	"context"
)

// Executor defines the interface for running shell commands
type Executor interface {
	// Execute runs a shell command and returns the output
	Execute(ctx context.Context, command string) (string, error)

	// ExecuteWithOutput runs a shell command and returns both stdout and stderr
	ExecuteWithOutput(ctx context.Context, command string) (stdout string, stderr string, err error)

	// ExecuteSilent runs a shell command without output (for side effects only)
	ExecuteSilent(ctx context.Context, command string) error
}
