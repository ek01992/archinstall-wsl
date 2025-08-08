package tx

import "testing"

func TestTransaction_RollbackLIFOAndStopOnFirstError(t *testing.T) {
	tr := New()
	order := []int{}
	tr.Defer(func() error { order = append(order, 1); return nil })
	err2 := false
	tr.Defer(func() error {
		order = append(order, 2)
		if !err2 {
			err2 = true
			return assertErr
		}
		return nil
	})
	tr.Defer(func() error { order = append(order, 3); return nil })

	if err := tr.Rollback(); err == nil {
		t.Fatalf("expected error on rollback")
	}
	// Order must be 3,2,1
	if len(order) != 3 || order[0] != 3 || order[1] != 2 || order[2] != 1 {
		t.Fatalf("unexpected rollback order: %v", order)
	}
}

func TestTransaction_CommitDiscardsUndos(t *testing.T) {
	tr := New()
	called := false
	tr.Defer(func() error { called = true; return nil })
	tr.Commit()
	if err := tr.Rollback(); err != nil {
		t.Fatalf("rollback after commit should be nil, got %v", err)
	}
	if called {
		t.Fatalf("undo should not have been called after commit")
	}
}

type errString string

func (e errString) Error() string { return string(e) }

const assertErr = errString("boom")
