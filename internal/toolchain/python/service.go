package python

import (
	"fmt"
	"regexp"
	"strings"

	"archwsl-tui-configurator/internal/logx"
)

type Runner interface {
	Run(name string, args ...string) error
	Output(name string, args ...string) (string, error)
}

type VersionSource interface{ LatestPython() (string, error) }

type Service struct {
	r  Runner
	vs VersionSource
}

func NewService(r Runner, vs VersionSource) *Service { return &Service{r: r, vs: vs} }

var rePy = regexp.MustCompile(`(?i)^Python\s+([0-9]+\.[0-9]+\.[0-9]+)`) // matches "Python 3.x.y"

func (s *Service) current() (string, error) {
	out, err := s.r.Output("python", "--version")
	if err != nil {
		return "", err
	}
	m := rePy.FindStringSubmatch(strings.TrimSpace(out))
	if len(m) < 2 {
		return "", fmt.Errorf("unable to parse python version output: %q", out)
	}
	return m[1], nil
}

func (s *Service) Install() error {
	logx.Info("toolchain: ensure python")
	latest, err := s.vs.LatestPython()
	if err != nil || strings.TrimSpace(latest) == "" {
		return fmt.Errorf("fetch latest python: %w", err)
	}
	if _, err := s.r.Output("pyenv", "--version"); err != nil {
		if err := s.r.Run("pacman", "-S", "--noconfirm", "pyenv"); err != nil {
			return fmt.Errorf("install pyenv: %w", err)
		}
	}
	cur, err := s.current()
	if err != nil || cur != latest {
		if err := s.r.Run("pyenv", "install", "-s", latest); err != nil {
			return fmt.Errorf("pyenv install %s: %w", latest, err)
		}
		if err := s.r.Run("pyenv", "global", latest); err != nil {
			return fmt.Errorf("pyenv global %s: %w", latest, err)
		}
		cur2, err := s.current()
		if err != nil {
			return fmt.Errorf("verify python after configure: %w", err)
		}
		if cur2 != latest {
			logx.Error("toolchain: python verify failed", "expected", latest, "got", cur2)
			return fmt.Errorf("verification failed: expected Python %s, got %s", latest, cur2)
		}
	}
	if _, err := s.r.Output("pipx", "--version"); err != nil {
		if err := s.r.Run("pacman", "-S", "--noconfirm", "pipx"); err != nil {
			return fmt.Errorf("install pipx: %w", err)
		}
		if _, err := s.r.Output("pipx", "--version"); err != nil {
			return fmt.Errorf("verification failed: pipx not available")
		}
	}
	return nil
}
