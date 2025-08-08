package user

import (
	stduser "os/user"
	"strings"
)

var (
	// assertErr is an arbitrary error used in tests to simulate lookup failure.
	assertErr error = stduser.UnknownUserError("unknown")
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
