package firewall

import (
	"errors"
	"strings"
	"testing"
)

func TestConfigureFirewallTx_RollsBackOnFailure(t *testing.T) {
	origRun := runCommand
	origCap := runCommandCapture
	t.Cleanup(func() { runCommand = origRun; runCommandCapture = origCap })

	// Start inactive
	runCommandCapture = func(name string, args ...string) (string, error) {
		if name == "ufw" && len(args) >= 1 && args[0] == "status" {
			return "Status: inactive\n", nil
		}
		return "", nil
	}

	var calls []string
	runCommand = func(name string, args ...string) error {
		calls = append(calls, strings.Join(append([]string{name}, args...), " "))
		// Induce failure when applying configuration
		if name == "ufw" && len(args) >= 1 && args[0] == "default" {
			return errors.New("fail")
		}
		return nil
	}

	if err := configureFirewallTx(); err == nil {
		t.Fatalf("expected error from configureFirewallTx")
	}
	// Expect rollback call to disable (since initial was inactive, undo tries disable)
	foundDisable := false
	for _, c := range calls {
		if strings.Contains(c, "ufw disable") {
			foundDisable = true
			break
		}
	}
	if !foundDisable {
		t.Fatalf("expected ufw disable in rollback; calls=%v", calls)
	}
}

func TestConfigureFirewallTx_SuccessDoesNotRollback(t *testing.T) {
	origRun := runCommand
	origCap := runCommandCapture
	t.Cleanup(func() { runCommand = origRun; runCommandCapture = origCap })

	runCommandCapture = func(name string, args ...string) (string, error) {
		if name == "ufw" && len(args) >= 1 && args[0] == "status" {
			return "Status: active\n", nil
		}
		return "", nil
	}

	var rolledBack bool
	runCommand = func(name string, args ...string) error {
		if name == "ufw" && len(args) >= 1 && (args[0] == "disable" || (len(args) >= 2 && args[0] == "--force" && args[1] == "enable")) {
			rolledBack = true
		}
		return nil
	}

	if err := configureFirewallTx(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rolledBack {
		t.Fatalf("did not expect rollback on success")
	}
}
