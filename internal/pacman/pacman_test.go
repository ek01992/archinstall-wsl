package pacman

import "testing"

func TestIsPackageInstalledWithQuery_EmptyFalse(t *testing.T) {
	if isPackageInstalledWithQuery("   ", func(string) error { return nil }) {
		t.Fatalf("empty name should be false")
	}
}

func TestIsPackageInstalledWithQuery_Success(t *testing.T) {
	if !isPackageInstalledWithQuery("bash", func(string) error { return nil }) {
		t.Fatalf("expected true on nil error")
	}
}

func TestIsPackageInstalledWithQuery_Error(t *testing.T) {
	if isPackageInstalledWithQuery("bash", func(string) error { return errStr("x") }) {
		t.Fatalf("expected false on error")
	}
}

type errStr string

func (e errStr) Error() string { return string(e) }
