package app

import "testing"

func TestNewReturnsApp(t *testing.T) {
	a := New()
	if a == nil {
		t.Fatalf("New() returned nil; expected non-nil *App")
	}
}
