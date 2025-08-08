package firewall

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

var (
	runCommand = func(name string, args ...string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, name, args...)
		return cmd.Run()
	}

	runCommandCapture = func(name string, args ...string) (string, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, name, args...)
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out
		if err := cmd.Run(); err != nil {
			return "", err
		}
		return out.String(), nil
	}
)

// configureFirewall ensures ufw is configured for default-deny inbound, allow outgoing,
// allows Windows↔WSL subnet access, and is enabled. It is idempotent.
func configureFirewall() error {
	status, err := runCommandCapture("ufw", "status", "verbose")
	if err != nil {
		// Fallback to plain status if verbose not supported, but keep the error otherwise
		if s2, err2 := runCommandCapture("ufw", "status"); err2 == nil {
			status = s2
		} else {
			return fmt.Errorf("ufw status failed: %w", err)
		}
	}

	isActive := strings.Contains(status, "Status: active")
	hasDenyIncoming := strings.Contains(status, "deny (incoming)")
	hasAllowOutgoing := strings.Contains(status, "allow (outgoing)")
	hasSubnetRule := strings.Contains(status, "172.16.0.0/12")

	// Set defaults if missing
	if !hasDenyIncoming {
		if err := runCommand("ufw", "default", "deny", "incoming"); err != nil {
			return fmt.Errorf("set default deny incoming: %w", err)
		}
	}
	if !hasAllowOutgoing {
		if err := runCommand("ufw", "default", "allow", "outgoing"); err != nil {
			return fmt.Errorf("set default allow outgoing: %w", err)
		}
	}

	// Allow Windows host subnet (WSL uses 172.16.0.0/12 range commonly for host↔WSL)
	if !hasSubnetRule {
		if err := runCommand("ufw", "allow", "from", "172.16.0.0/12"); err != nil {
			return fmt.Errorf("allow subnet: %w", err)
		}
	}

	// Enable ufw if inactive
	if !isActive {
		// Use --force to avoid interactive prompt
		if err := runCommand("ufw", "--force", "enable"); err != nil {
			// Some ufw versions may not support --force; try plain enable
			if !errors.Is(err, context.DeadlineExceeded) {
				if err2 := runCommand("ufw", "enable"); err2 == nil {
					return nil
				}
			}
			return fmt.Errorf("enable ufw: %w", err)
		}
	}

	return nil
}
