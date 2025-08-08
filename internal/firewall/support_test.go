package firewall

import (
	"errors"
	"testing"
)

func TestIsUfwSupported_TrueWhenStatusOk(t *testing.T) {
	orig := runCommandCapture
	defer func() { runCommandCapture = orig }()
	runCommandCapture = func(name string, args ...string) (string, error) { return "Status: active", nil }
	if !isUfwSupported() {
		t.Fatalf("expected true when ufw status returns output")
	}
}

func TestIsUfwSupported_FalseOnError(t *testing.T) {
	orig := runCommandCapture
	defer func() { runCommandCapture = orig }()
	runCommandCapture = func(name string, args ...string) (string, error) { return "", errors.New("no ufw") }
	if isUfwSupported() {
		t.Fatalf("expected false when ufw status errors")
	}
}
