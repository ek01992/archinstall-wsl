package firewall

import (
	"strings"

	"archwsl-tui-configurator/internal/tx"
)

// configureFirewallTx ensures firewall configuration with rollback on failure.
func configureFirewallTx() (err error) {
	tr := tx.New()
	defer func() { if err != nil { _ = tr.Rollback() } }()

	// Capture initial status output to detect whether ufw was active
	status, _ := runCommandCapture("ufw", "status")
	wasActive := strings.Contains(status, "Status: active")
	if wasActive {
		tr.Defer(func() error { return runCommand("ufw", "--force", "enable") })
	} else {
		tr.Defer(func() error { return runCommand("ufw", "disable") })
	}

	// We cannot easily diff individual defaults/rules here; rely on the disable/enable reversal
	if err = configureFirewall(); err != nil {
		return err
	}
	tr.Commit()
	return nil
}
