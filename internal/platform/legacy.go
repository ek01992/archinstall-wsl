package platform

import (
	"os"
	"time"
)

// OS-backed adapters to avoid recursion when redirecting seams

type osFS struct{}

func (osFS) ReadFile(path string) ([]byte, error)  { return os.ReadFile(path) }
func (osFS) Stat(path string) (interface{}, error) { return os.Stat(path) }

type osEnv struct{}

func (osEnv) Getenv(k string) string { return os.Getenv(k) }

var defaultService = NewService(osFS{}, osEnv{})

// Redirect legacy seams to the DI service safely
func init() {
	getenv = func(k string) string { return defaultService.Getenv(k) }
	pathExists = func(path string) bool { return defaultService.IsMounted(path) }
	readFile = func(path string) ([]byte, error) { return defaultService.fs.ReadFile(path) }
}

var _ = time.Second
