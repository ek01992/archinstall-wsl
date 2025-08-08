package user

import "testing"

func TestGetDefaultShell_ReturnsShellWhenUserExists(t *testing.T) {
	orig := readPasswdFileBytes
	t.Cleanup(func() { readPasswdFileBytes = orig })

	readPasswdFileBytes = func() ([]byte, error) {
		return []byte(`# Comment line
root:x:0:0:root:/root:/bin/bash
alice:x:1000:1000:Alice:/home/alice:/usr/bin/zsh
`), nil
	}

	shell := getDefaultShell("alice")
	if shell != "/usr/bin/zsh" {
		t.Fatalf("expected /usr/bin/zsh, got %q", shell)
	}
}

func TestGetDefaultShell_EmptyOrWhitespaceUsernameReturnsEmpty(t *testing.T) {
	orig := readPasswdFileBytes
	t.Cleanup(func() { readPasswdFileBytes = orig })

	called := false
	readPasswdFileBytes = func() ([]byte, error) { called = true; return nil, nil }

	if got := getDefaultShell("  \t\n"); got != "" {
		t.Fatalf("expected empty string, got %q", got)
	}
	if called {
		t.Fatalf("passwd file should not be read for empty/whitespace username")
	}
}

func TestGetDefaultShell_UserNotFoundReturnsEmpty(t *testing.T) {
	orig := readPasswdFileBytes
	t.Cleanup(func() { readPasswdFileBytes = orig })

	readPasswdFileBytes = func() ([]byte, error) {
		return []byte(`root:x:0:0:root:/root:/bin/bash
bob:x:1001:1001:Bob:/home/bob:/bin/fish
`), nil
	}

	if got := getDefaultShell("charlie"); got != "" {
		t.Fatalf("expected empty string for missing user, got %q", got)
	}
}

func TestGetDefaultShell_IgnoresMalformedAndTrimsShell(t *testing.T) {
	orig := readPasswdFileBytes
	t.Cleanup(func() { readPasswdFileBytes = orig })

	readPasswdFileBytes = func() ([]byte, error) {
		return []byte(`# bad line
malformed
frank:x:1002:1002::/home/frank:   /bin/bash   
`), nil
	}

	if got := getDefaultShell("frank"); got != "/bin/bash" {
		t.Fatalf("expected /bin/bash, got %q", got)
	}
}
