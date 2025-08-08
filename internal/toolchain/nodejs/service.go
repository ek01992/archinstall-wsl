package nodejs

import (
	"fmt"
	"regexp"
	"strings"
)

type Runner interface {
	Run(name string, args ...string) error
	Output(name string, args ...string) (string, error)
	Shell(cmd string) (string, error)
}

type VersionSource interface { LatestLTS() (string, error) }

type Service struct { r Runner; vs VersionSource }

func NewService(r Runner, vs VersionSource) *Service { return &Service{r: r, vs: vs} }

var reNode = regexp.MustCompile(`^v\d+\.\d+\.\d+$`)

func (s *Service) current() (string, error) {
	out, err := s.r.Output("node", "-v")
	if err != nil {
		if out2, err2 := s.r.Shell("node -v"); err2 == nil { out = out2 } else { return "", err }
	}
	ver := strings.TrimSpace(out)
	if !reNode.MatchString(ver) { return "", fmt.Errorf("unable to parse node -v output: %q", ver) }
	return ver, nil
}

func (s *Service) Install() error {
	lts, err := s.vs.LatestLTS()
	if err != nil || strings.TrimSpace(lts) == "" { return fmt.Errorf("fetch latest node LTS: %w", err) }
	if _, err := s.r.Output("nvm", "--version"); err != nil {
		if err := s.r.Run("pacman", "-S", "--noconfirm", "nvm"); err != nil { return fmt.Errorf("install nvm: %w", err) }
	}
	cur, err := s.current()
	if err != nil || cur != lts {
		if err := s.r.Run("nvm", "install", lts); err != nil {
			if _, err2 := s.r.Shell("source /usr/share/nvm/init-nvm.sh 2>/dev/null || true; nvm install "+lts); err2 != nil {
				return fmt.Errorf("nvm install %s: %w", lts, err)
			}
		}
		if err := s.r.Run("nvm", "alias", "default", lts); err != nil {
			if _, err2 := s.r.Shell("source /usr/share/nvm/init-nvm.sh 2>/dev/null || true; nvm alias default "+lts); err2 != nil {
				return fmt.Errorf("nvm alias default %s: %w", lts, err)
			}
		}
		nv, err := s.current(); if err != nil { return fmt.Errorf("verify node after install: %w", err) }
		if nv != lts { return fmt.Errorf("verification failed: expected %s, got %s", lts, nv) }
	}
	return nil
}
