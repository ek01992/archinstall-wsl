package seams

import (
	"sync/atomic"
	"testing"
)

func TestWith_SerializesOverrides(t *testing.T) {
	var target int32 = 1
	var saw int32

	With(&target, 2, func() {
		atomic.StoreInt32(&saw, target)
	})

	if atomic.LoadInt32(&saw) != 2 {
		t.Fatalf("expected override to be visible within With block")
	}
	if target != 1 {
		t.Fatalf("expected target restored after With; got %d", target)
	}
}
