package golang

import (
	"fmt"
	"regexp"
	"strings"
)

type Runner interface {
	Run(name string, args ...string) error
	Output(name string, args ...string) (string, error)
}

type VersionSource interface { LatestGo() (string, error) }

type Service struct { r Runner; vs VersionSource }

func NewService(r Runner, vs VersionSource) *Service { return &Service{r: r, vs: vs} }

var reGo = regexp.MustCompile(`go version go([0-9]+\.[0-9]+\.[0-9]+) `)

func (s *Service) current() (string, error) {
	out, err := s.r.Output("go", "version")
	if err != nil { return "", err }
	m := reGo.FindStringSubmatch(out)
	if len(m) < 2 { return "", fmt.Errorf("unable to parse go version output: %q", out) }
	return m[1], nil
}

func (s *Service) Install() error {
	latest, err := s.vs.LatestGo()
	if err != nil || strings.TrimSpace(latest) == "" { return fmt.Errorf("fetch latest go version: %w", err) }
	cur, err := s.current()
	if err != nil {
		if err := s.r.Run("pacman", "-S", "--noconfirm", "go"); err != nil { return fmt.Errorf("install go: %w", err) }
		if _, err := s.current(); err != nil { return fmt.Errorf("verify go after install: %w", err) }
		return nil
	}
	if cur == latest { return nil }
	if err := s.r.Run("pacman", "-Syu", "--noconfirm", "go"); err != nil { return fmt.Errorf("update go: %w", err) }
	cur2, err := s.current()
	if err != nil { return fmt.Errorf("verify go after update: %w", err) }
	if cur2 != latest { return fmt.Errorf("verification failed: expected %s, got %s", latest, cur2) }
	return nil
}
