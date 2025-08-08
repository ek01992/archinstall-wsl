package pacman

import (
	"errors"
	"testing"
)

func TestIsPackageInstalled_TrueOnSuccessfulQuery(t *testing.T) {
	orig := queryLocalPackage
	t.Cleanup(func() { queryLocalPackage = orig })

	called := false
	queryLocalPackage = func(pkg string) error {
		called = true
		if pkg != "git" {
			t.Fatalf("expected pkg 'git', got %q", pkg)
		}
		return nil
	}

	if !isPackageInstalled("git") {
		t.Fatalf("expected true when query succeeds")
	}
	if !called {
		t.Fatalf("expected queryLocalPackage to be called")
	}
}

func TestIsPackageInstalled_FalseOnError(t *testing.T) {
	orig := queryLocalPackage
	t.Cleanup(func() { queryLocalPackage = orig })

	queryLocalPackage = func(pkg string) error { return errors.New("not installed") }
	if isPackageInstalled("vim") {
		t.Fatalf("expected false when query returns error")
	}
}

func TestIsPackageInstalled_EmptyStringFalse(t *testing.T) {
	orig := queryLocalPackage
	t.Cleanup(func() { queryLocalPackage = orig })

	called := false
	queryLocalPackage = func(pkg string) error { called = true; return nil }
	if isPackageInstalled("   \t\n") {
		t.Fatalf("expected false for empty/whitespace package name")
	}
	if called {
		t.Fatalf("query should not be called for empty input")
	}
}
