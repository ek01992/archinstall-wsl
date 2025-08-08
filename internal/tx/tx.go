package tx

import "log"

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

// Rollback executes registered undo actions in reverse order and returns an aggregated error if any occurred.
func (t *Transaction) Rollback() error {
	if t == nil || t.committed {
		return nil
	}
	var agg multiError
	for i := len(t.undos) - 1; i >= 0; i-- {
		if err := t.undos[i](); err != nil {
			// Log every rollback error for observability
			log.Printf("tx rollback error: %v", err)
			agg.append(err)
		}
	}
	// Clear after rollback
	t.undos = nil
	if agg.len() == 0 {
		return nil
	}
	return agg
}

type multiError struct{ errs []error }

func (m *multiError) append(err error) { m.errs = append(m.errs, err) }
func (m *multiError) len() int         { return len(m.errs) }

func (m multiError) Error() string {
	if len(m.errs) == 1 {
		return m.errs[0].Error()
	}
	msg := "rollback encountered multiple errors:"
	for _, e := range m.errs {
		msg += " " + e.Error() + ";"
	}
	return msg
}
