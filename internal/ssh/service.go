package ssh

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

type Platform interface { CanEditHostFiles() bool }

type FS interface {
	ReadDir(dir string) ([]string, error)
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte, perm fs.FileMode) error
	MkdirAll(path string, perm fs.FileMode) error
	Chmod(path string, mode fs.FileMode) error
	UserHomeDir() (string, error)
}

type Service struct { p Platform; fs FS }

func NewService(p Platform, fs FS) *Service { return &Service{p: p, fs: fs} }

func (s *Service) ImportWithConsent(hostPath string, consent bool) error {
	if !consent { return fmt.Errorf("ssh key import: explicit consent required") }
	if !s.p.CanEditHostFiles() { return fmt.Errorf("ssh key import: host files not accessible (WSL mount missing)") }
	return s.ImportFromWindows(hostPath)
}

func (s *Service) ImportFromWindows(hostPath string) error {
	hostPath = strings.TrimSpace(hostPath)
	if hostPath == "" { return errors.New("hostPath must not be empty") }
	home, err := s.fs.UserHomeDir(); if err != nil { return fmt.Errorf("resolve home dir: %w", err) }
	if strings.TrimSpace(home) == "" { return errors.New("empty home directory") }
	dotSSH := filepath.Join(home, ".ssh")
	if err := s.fs.MkdirAll(dotSSH, 0o700); err != nil { return fmt.Errorf("ensure ~/.ssh: %w", err) }
	_ = s.fs.Chmod(dotSSH, 0o700)
	names, err := s.fs.ReadDir(hostPath); if err != nil { return fmt.Errorf("list host ssh dir: %w", err) }
	for _, name := range names {
		if name == "" || name == "." || name == ".." { continue }
		src := filepath.Join(hostPath, name)
		dst := filepath.Join(dotSSH, name)
		srcBytes, err := s.fs.ReadFile(src); if err != nil { return fmt.Errorf("read source %s: %w", src, err) }
		mode := fs.FileMode(0o600)
		if strings.HasSuffix(name, ".pub") { mode = 0o644 }
		dstBytes, err := s.fs.ReadFile(dst)
		if err == nil && bytes.Equal(dstBytes, srcBytes) { _ = s.fs.Chmod(dst, mode); continue }
		if err := s.fs.WriteFile(dst, srcBytes, mode); err != nil { return fmt.Errorf("write destination %s: %w", dst, err) }
		_ = s.fs.Chmod(dst, mode)
	}
	return nil
}
