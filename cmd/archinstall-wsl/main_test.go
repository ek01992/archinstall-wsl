package main

import (
	"testing"
)

func TestMain(t *testing.T) {
	// This test verifies that the main package can be imported and compiled
	// The actual TUI functionality is tested in the internal/tui package
	t.Log("Main package compiles successfully")
}
