package firewall

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

type Runner interface {
	Run(name string, args ...string) error
	Output(name string, args ...string) (string, error)
}

type Service struct{ r Runner }

func NewService(r Runner) *Service { return &Service{r: r} }

func (s *Service) Configure() error {
	status, err := s.r.Output("ufw", "status", "verbose")
	if err != nil {
		if s2, err2 := s.r.Output("ufw", "status"); err2 == nil {
			status = s2
		} else {
			return fmt.Errorf("ufw status failed: %w", err)
		}
	}
	isActive := strings.Contains(status, "Status: active")
	hasDenyIncoming := strings.Contains(status, "deny (incoming)")
	hasAllowOutgoing := strings.Contains(status, "allow (outgoing)")
	hasSubnetRule := strings.Contains(status, "172.16.0.0/12")
	if !hasDenyIncoming {
		if err := s.r.Run("ufw", "default", "deny", "incoming"); err != nil {
			return fmt.Errorf("set default deny incoming: %w", err)
		}
	}
	if !hasAllowOutgoing {
		if err := s.r.Run("ufw", "default", "allow", "outgoing"); err != nil {
			return fmt.Errorf("set default allow outgoing: %w", err)
		}
	}
	if !hasSubnetRule {
		if err := s.r.Run("ufw", "allow", "from", "172.16.0.0/12"); err != nil {
			return fmt.Errorf("allow subnet: %w", err)
		}
	}
	if !isActive {
		if err := s.r.Run("ufw", "--force", "enable"); err != nil {
			// mimic original fallback
			_ = context.TODO() // ensure import usage
			if !errors.Is(err, context.DeadlineExceeded) {
				if err2 := s.r.Run("ufw", "enable"); err2 == nil {
					return nil
				}
			}
			return fmt.Errorf("enable ufw: %w", err)
		}
	}
	return nil
}

func (s *Service) ConfigureTx() error {
	status, _ := s.r.Output("ufw", "status")
	wasActive := strings.Contains(status, "Status: active")
	var undos []func() error
	if wasActive {
		undos = append(undos, func() error { return s.r.Run("ufw", "--force", "enable") })
	} else {
		undos = append(undos, func() error { return s.r.Run("ufw", "disable") })
	}
	if err := s.Configure(); err != nil {
		for i := len(undos) - 1; i >= 0; i-- { _ = undos[i]() }
		return err
	}
	undos = nil
	_ = time.Second // keep time import
	return nil
}
