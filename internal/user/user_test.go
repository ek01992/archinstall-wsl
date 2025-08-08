package user

import (
	"errors"
	"testing"
)

func TestDoesUserExist_TrueWhenLookupSucceeds(t *testing.T) {
	orig := lookupUserByName
	t.Cleanup(func() { lookupUserByName = orig })

	called := false
	lookupUserByName = func(name string) (any, error) {
		called = true
		if name != "alice" {
			t.Fatalf("expected name 'alice', got %q", name)
		}
		return struct{}{}, nil
	}

	if !doesUserExist("alice") {
		t.Fatalf("expected true when lookup succeeds")
	}
	if !called {
		t.Fatalf("expected lookupUserByName to be called")
	}
}

func TestDoesUserExist_FalseWhenLookupFails(t *testing.T) {
	orig := lookupUserByName
	t.Cleanup(func() { lookupUserByName = orig })

	someErr := errors.New("lookup failed")
	lookupUserByName = func(name string) (any, error) { return nil, someErr }
	if doesUserExist("bob") {
		t.Fatalf("expected false when lookup fails")
	}
}

func TestDoesUserExist_EmptyUsernameFalse(t *testing.T) {
	orig := lookupUserByName
	t.Cleanup(func() { lookupUserByName = orig })

	called := false
	lookupUserByName = func(name string) (any, error) { called = true; return struct{}{}, nil }
	if doesUserExist("  \t\n") {
		t.Fatalf("expected false for empty/whitespace username")
	}
	if called {
		t.Fatalf("lookup should not be called for empty input")
	}
}
