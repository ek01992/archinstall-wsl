package platform

import (
	"os"
	"strings"
)

// NOTE: Package-level seams below are for testability and are NOT concurrency-safe.
// Use internal/seams.With in tests to serialize overrides. Prefer DI if adding concurrency.
var (
	readFile   = func(path string) ([]byte, error) { return os.ReadFile(path) }
	pathExists = func(path string) bool {
		_, err := os.Stat(path)
		return err == nil
	}
	getenv = func(k string) string { return os.Getenv(k) }
)

// IsWSL returns true if the current environment appears to be WSL.
func IsWSL() bool {
	if getenv("WSL_INTEROP") != "" || getenv("WSL_DISTRO_NAME") != "" {
		return true
	}
	if b, err := readFile("/proc/sys/kernel/osrelease"); err == nil && strings.Contains(strings.ToLower(string(b)), "microsoft") {
		return true
	}
	if b, err := readFile("/proc/version"); err == nil && strings.Contains(strings.ToLower(string(b)), "microsoft") {
		return true
	}
	return false
}

// IsMounted returns true if the given path exists (best-effort).
func IsMounted(path string) bool { return pathExists(path) }

// CanEditHostFiles returns true if running under WSL and Windows drive C: is mounted.
func CanEditHostFiles() bool { return IsWSL() && IsMounted("/mnt/c") }
