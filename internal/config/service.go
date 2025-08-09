package config

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type FS interface {
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte, perm fs.FileMode) error
	MkdirAll(path string, perm fs.FileMode) error
}

type Service struct {
	fs   FS
	home func() (string, error)
}

func NewService(fs FS, home func() (string, error)) *Service { return &Service{fs: fs, home: home} }

func (s *Service) Save(cfg Config) error {
	home, err := s.home()
	if err != nil || strings.TrimSpace(home) == "" {
		return fmt.Errorf("cannot determine home directory: %w", err)
	}
	dir := filepath.Join(home, ".config", "archwsl-tui-configurator")
	if err := s.fs.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("mkdir config dir: %w", err)
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal yaml: %w", err)
	}
	path := filepath.Join(dir, "config.yaml")
	if err := s.fs.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

func (s *Service) Load(path string) Config {
	p := strings.TrimSpace(path)
	if p == "" {
		home, err := s.home()
		if err != nil || strings.TrimSpace(home) == "" {
			return Config{}
		}
		p = filepath.Join(home, ".config", "archwsl-tui-configurator", "config.yaml")
	}
	b, err := s.fs.ReadFile(p)
	if err != nil {
		return Config{}
	}
	var cfg Config
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return Config{}
	}
	return cfg
}
