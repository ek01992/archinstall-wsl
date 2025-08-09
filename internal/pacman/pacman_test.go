package pacman

import (
	"errors"
	"testing"
)

func TestIsPackageInstalled_TrueOnSuccessfulQuery(t *testing.T) {
	called := false
	q := func(pkg string) error {
		called = true
		if pkg != "git" { t.Fatalf("expected pkg 'git', got %q", pkg) }
		return nil
	}
	if !isPackageInstalledWithQuery("git", q) { t.Fatalf("expected true when query succeeds") }
	if !called { t.Fatalf("expected query to be called") }
}

func TestIsPackageInstalled_FalseOnError(t *testing.T) {
	q := func(pkg string) error { return errors.New("not installed") }
	if isPackageInstalledWithQuery("vim", q) { t.Fatalf("expected false when query returns error") }
}

func TestIsPackageInstalled_EmptyStringFalse(t *testing.T) {
	called := false
	q := func(pkg string) error { called = true; return nil }
	if isPackageInstalledWithQuery("   \t\n", q) { t.Fatalf("expected false for empty/whitespace package name") }
	if called { t.Fatalf("query should not be called for empty input") }
}

func TestIsPackageInstalled_DefaultFastPath(t *testing.T) {
	// Calling default function with empty input should not spawn pacman and returns false
	if isPackageInstalled("   ") { t.Fatalf("expected false for empty input") }
}
