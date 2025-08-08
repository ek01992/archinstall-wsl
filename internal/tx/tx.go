package tx

// Action represents a compensating action to revert a prior change.
type Action func() error

// Transaction records compensating actions to rollback on failure.
type Transaction struct {
	undos     []Action
	committed bool
}

// New creates an empty transaction.
func New() *Transaction { return &Transaction{undos: make([]Action, 0, 8)} }

// Defer registers an undo action to run on rollback in LIFO order.
func (t *Transaction) Defer(a Action) {
	if t == nil || t.committed || a == nil {
		return
	}
	t.undos = append(t.undos, a)
}

// Commit marks the transaction as successful; discards undo actions.
func (t *Transaction) Commit() { t.committed = true; t.undos = nil }

// Rollback executes registered undo actions in reverse order and returns the first error encountered.
func (t *Transaction) Rollback() error {
	if t == nil || t.committed {
		return nil
	}
	var firstErr error
	for i := len(t.undos) - 1; i >= 0; i-- {
		if err := t.undos[i](); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	// Clear after rollback
	t.undos = nil
	return firstErr
}
