package platform

import "time"

// Seam-backed adapters to preserve test overrides

type seamFS struct{}

func (seamFS) ReadFile(path string) ([]byte, error) { return readFile(path) }
func (seamFS) Stat(path string) (interface{}, error) {
	if pathExists(path) {
		return struct{}{}, nil
	}
	return nil, osErrNotExist
}

// Minimal error type to avoid importing os; compare by nil/non-nil
var osErrNotExist = errString("not-exist")

type errString string

func (e errString) Error() string { return string(e) }

type seamEnv struct{}

func (seamEnv) Getenv(k string) string { return getenv(k) }

// Default DI-backed service available for new callers
var defaultService = NewService(seamFS{}, seamEnv{})

var _ = time.Second
