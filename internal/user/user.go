package user

import (
	stduser "os/user"
	"strings"
)

// NOTE: Package-level seams are for testability and are NOT concurrency-safe.
// Use internal/seams.With in tests to serialize overrides. Prefer DI if adding concurrency.
var (
	lookupUserByName = func(name string) (any, error) {
		_, err := stduser.Lookup(name)
		return struct{}{}, err
	}
)

// doesUserExist returns true if a local user with the given username exists.
// It trims whitespace and handles empty input by returning false.
func doesUserExist(username string) bool {
	username = strings.TrimSpace(username)
	if username == "" {
		return false
	}
	_, err := lookupUserByName(username)
	return err == nil
}
