package firewall

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"archwsl-tui-configurator/internal/logx"
)

type Runner interface {
	Run(name string, args ...string) error
	Output(name string, args ...string) (string, error)
}

type Service struct{ r Runner }

func NewService(r Runner) *Service { return &Service{r: r} }

func (s *Service) Configure() error {
	logx.Info("ufw: ensure defaults and rules")
	status, err := s.r.Output("ufw", "status", "verbose")
	if err != nil {
		if s2, err2 := s.r.Output("ufw", "status"); err2 == nil {
			status = s2
		} else {
			logx.Error("ufw: status failed", "err", err)
			return fmt.Errorf("ufw status failed: %w", err)
		}
		logx.Warn("ufw: verbose not available; falling back")
	}
	isActive := strings.Contains(status, "Status: active")
	hasDenyIncoming := strings.Contains(status, "deny (incoming)")
	hasAllowOutgoing := strings.Contains(status, "allow (outgoing)")
	hasSubnetRule := strings.Contains(status, "172.16.0.0/12")
	if !hasDenyIncoming {
		if err := s.r.Run("ufw", "default", "deny", "incoming"); err != nil {
			logx.Error("ufw: set default deny incoming failed", "err", err)
			return fmt.Errorf("set default deny incoming: %w", err)
		}
	}
	if !hasAllowOutgoing {
		if err := s.r.Run("ufw", "default", "allow", "outgoing"); err != nil {
			logx.Error("ufw: set default allow outgoing failed", "err", err)
			return fmt.Errorf("set default allow outgoing: %w", err)
		}
	}
	if !hasSubnetRule {
		if err := s.r.Run("ufw", "allow", "from", "172.16.0.0/12"); err != nil {
			logx.Error("ufw: allow subnet failed", "err", err)
			return fmt.Errorf("allow subnet: %w", err)
		}
	}
	if !isActive {
		if err := s.r.Run("ufw", "--force", "enable"); err != nil {
			_ = context.TODO()
			if !errors.Is(err, context.DeadlineExceeded) {
				if err2 := s.r.Run("ufw", "enable"); err2 == nil {
					return nil
				}
			}
			logx.Error("ufw: enable failed", "err", err)
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
		for i := len(undos) - 1; i >= 0; i-- {
			_ = undos[i]()
		}
		return err
	}
	_ = time.Second
	return nil
}
