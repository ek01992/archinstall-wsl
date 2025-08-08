package git

import (
	"fmt"
	"strings"
)

type Runner interface {
	Run(name string, args ...string) error
	Output(name string, args ...string) (string, error)
}

type Service struct{ r Runner }

func NewService(r Runner) *Service { return &Service{r: r} }

func (s *Service) Configure(userName, userEmail string) error {
	name := strings.TrimSpace(userName)
	email := strings.TrimSpace(userEmail)
	if name == "" || email == "" {
		return fmt.Errorf("name and email must not be empty")
	}
	if err := s.r.Run("git", "config", "--global", "user.name", name); err != nil {
		return fmt.Errorf("git config user.name failed: %w", err)
	}
	if err := s.r.Run("git", "config", "--global", "user.email", email); err != nil {
		return fmt.Errorf("git config user.email failed: %w", err)
	}
	gotName, err := s.r.Output("git", "config", "--global", "--get", "user.name")
	if err != nil {
		return fmt.Errorf("verify user.name failed: %w", err)
	}
	gotEmail, err := s.r.Output("git", "config", "--global", "--get", "user.email")
	if err != nil {
		return fmt.Errorf("verify user.email failed: %w", err)
	}
	if strings.TrimSpace(gotName) != name || strings.TrimSpace(gotEmail) != email {
		return fmt.Errorf("verification failed: expected name %q and email %q, got %q / %q",
			name, email, strings.TrimSpace(gotName), strings.TrimSpace(gotEmail))
	}
	return nil
}

func (s *Service) ConfigureTx(userName, userEmail string) error {
	prevName, _ := s.r.Output("git", "config", "--global", "--get", "user.name")
	prevEmail, _ := s.r.Output("git", "config", "--global", "--get", "user.email")
	prevName = strings.TrimSpace(prevName)
	prevEmail = strings.TrimSpace(prevEmail)

	type undo func() error
	var undos []undo
	undos = append(undos, func() error {
		if prevName == "" {
			_ = s.r.Run("git", "config", "--global", "--unset", "user.name")
		} else {
			_ = s.r.Run("git", "config", "--global", "user.name", prevName)
		}
		if prevEmail == "" {
			_ = s.r.Run("git", "config", "--global", "--unset", "user.email")
		} else {
			_ = s.r.Run("git", "config", "--global", "user.email", prevEmail)
		}
		return nil
	})

	if err := s.Configure(userName, userEmail); err != nil {
		// rollback
		for i := len(undos) - 1; i >= 0; i-- {
			_ = undos[i]()
		}
		return err
	}
	// commit
	return nil
}
