package seams

import "sync"

var mu sync.Mutex

// With temporarily sets the target variable to temp, runs fn, then restores the original value.
// It guards the override with a global mutex to avoid concurrent races in tests.
func With[T any](target *T, temp T, fn func()) {
	mu.Lock()
	orig := *target
	*target = temp
	defer func() {
		*target = orig
		mu.Unlock()
	}()
	fn()
}
