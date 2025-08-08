package firewall

import (
	"errors"
	"strings"
	"testing"
)

func TestConfigureFirewall_SetsDefaults_AllowsSubnet_Enables(t *testing.T) {
	origRun := runCommand
	origCap := runCommandCapture
	t.Cleanup(func() {
		runCommand = origRun
		runCommandCapture = origCap
	})

	// First, status shows inactive and no defaults
	statusOutput := "Status: inactive\n"

	runCommandCapture = func(name string, args ...string) (string, error) {
		if name != "ufw" {
			t.Fatalf("expected ufw, got %q", name)
		}
		if len(args) >= 1 && args[0] == "status" {
			if len(args) >= 2 && args[1] == "verbose" {
				return statusOutput, nil
			}
			return "Status: inactive\n", nil
		}
		return "", errors.New("unexpected capture call")
	}

	var setDenyIncoming, setAllowOutgoing, allowedSubnet, enabled bool

	runCommand = func(name string, args ...string) error {
		if name != "ufw" {
			t.Fatalf("expected ufw, got %q", name)
		}
		a := strings.Join(args, " ")
		switch {
		case strings.HasPrefix(a, "default deny incoming"):
			setDenyIncoming = true
			return nil
		case strings.HasPrefix(a, "default allow outgoing"):
			setAllowOutgoing = true
			return nil
		case strings.HasPrefix(a, "allow from 172.16.0.0/12"):
			allowedSubnet = true
			return nil
		case strings.HasPrefix(a, "--force enable") || strings.HasPrefix(a, "enable"):
			enabled = true
			return nil
		default:
			t.Fatalf("unexpected ufw args: %v", args)
			return nil
		}
	}

	if err := configureFirewall(); err != nil {
		t.Fatalf("configureFirewall returned error: %v", err)
	}

	if !setDenyIncoming {
		t.Fatalf("expected to set default deny incoming")
	}
	if !setAllowOutgoing {
		t.Fatalf("expected to set default allow outgoing")
	}
	if !allowedSubnet {
		t.Fatalf("expected to allow 172.16.0.0/12 subnet")
	}
	if !enabled {
		t.Fatalf("expected ufw to be enabled")
	}
}

func TestConfigureFirewall_Idempotent_WhenAlreadyConfigured(t *testing.T) {
	origRun := runCommand
	origCap := runCommandCapture
	t.Cleanup(func() {
		runCommand = origRun
		runCommandCapture = origCap
	})

	// Status already active, defaults set, and subnet allowed
	statusOutput := "Status: active\nDefault: deny (incoming), allow (outgoing), disabled (routed)\nTo                         Action      From\n--                         ------      ----\nAnywhere                    ALLOW       172.16.0.0/12\n"

	runCommandCapture = func(name string, args ...string) (string, error) {
		if name != "ufw" {
			t.Fatalf("expected ufw, got %q", name)
		}
		if len(args) >= 1 && args[0] == "status" {
			return statusOutput, nil
		}
		return "", errors.New("unexpected capture call")
	}

	calledUnexpected := false
	runCommand = func(name string, args ...string) error {
		calledUnexpected = true
		t.Fatalf("did not expect any state-changing ufw calls; got %v", append([]string{name}, args...))
		return nil
	}

	if err := configureFirewall(); err != nil {
		t.Fatalf("configureFirewall returned error: %v", err)
	}
	if calledUnexpected {
		t.Fatalf("idempotency violated: state-changing calls made while already configured")
	}
}
