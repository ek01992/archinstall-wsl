package system

import (
	"testing"
)

func TestSystemInterfaces(t *testing.T) {
	// Test that the interfaces can be used
	t.Run("Executor interface can be implemented", func(t *testing.T) {
		var _ Executor = (*MockExecutor)(nil)
	})

	t.Run("StateChecker interface can be implemented", func(t *testing.T) {
		var _ StateChecker = (*MockStateChecker)(nil)
	})
}
