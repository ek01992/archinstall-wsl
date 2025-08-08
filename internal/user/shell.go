package user

import (
	"os"
	"strings"
)

// readPasswdFileBytes is a test seam to allow providing fixture content.
var readPasswdFileBytes = func() ([]byte, error) {
	return os.ReadFile("/etc/passwd")
}

// getDefaultShell returns the user's login shell as defined in /etc/passwd.
// Returns empty string if the user does not exist, input is empty/whitespace,
// or the entry is malformed.
func getDefaultShell(username string) string {
	username = strings.TrimSpace(username)
	if username == "" {
		return ""
	}

	data, err := readPasswdFileBytes()
	if err != nil {
		return ""
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, ":")
		if len(parts) < 7 {
			continue
		}
		name := parts[0]
		if name != username {
			continue
		}
		shell := strings.TrimSpace(parts[6])
		return shell
	}
	return ""
}
