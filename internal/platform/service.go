package platform

import (
	"strings"
)

type FS interface {
	ReadFile(path string) ([]byte, error)
	Stat(path string) (interface{}, error) // We only need existence; legacy uses bool
}

type Env interface { Getenv(string) string }

type Service struct { fs FS; env Env }

func NewService(fs FS, env Env) *Service { return &Service{fs: fs, env: env} }

func (s *Service) Getenv(k string) string { return s.env.Getenv(k) }

func (s *Service) IsMounted(path string) bool {
	// Use Stat existence check
	if _, err := s.fs.Stat(path); err == nil { return true }
	return false
}

func (s *Service) IsWSL() bool {
	if s.Getenv("WSL_INTEROP") != "" || s.Getenv("WSL_DISTRO_NAME") != "" { return true }
	if b, err := s.fs.ReadFile("/proc/sys/kernel/osrelease"); err == nil && strings.Contains(strings.ToLower(string(b)), "microsoft") { return true }
	if b, err := s.fs.ReadFile("/proc/version"); err == nil && strings.Contains(strings.ToLower(string(b)), "microsoft") { return true }
	return false
}

func (s *Service) CanEditHostFiles() bool { return s.IsWSL() && s.IsMounted("/mnt/c") }
