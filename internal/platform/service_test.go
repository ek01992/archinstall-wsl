package platform

import (
	"errors"
	"testing"
)

type tfs struct{ files map[string][]byte; mounts map[string]bool }
func (t tfs) ReadFile(path string) ([]byte, error) { if b, ok := t.files[path]; ok { return b, nil }; return nil, errors.New("no file") }
func (t tfs) Stat(path string) (interface{}, error) { if t.mounts[path] { return struct{}{}, nil }; return nil, errors.New("nope") }

type tenv map[string]string
func (e tenv) Getenv(k string) string { return e[k] }

func TestService_IsWSL_ByEnv(t *testing.T) {
	s := NewService(tfs{}, tenv{"WSL_INTEROP": "1"})
	if !s.IsWSL() { t.Fatalf("expected IsWSL true when env set") }
}

func TestService_IsWSL_ByProc(t *testing.T) {
	s := NewService(tfs{files: map[string][]byte{"/proc/sys/kernel/osrelease": []byte("5.15.0-microsoft-standard-WSL2")}}, tenv{})
	if !s.IsWSL() { t.Fatalf("expected IsWSL true when osrelease mentions microsoft") }
}

func TestService_IsMounted_And_CanEditHostFiles(t *testing.T) {
	s := NewService(tfs{mounts: map[string]bool{"/mnt/c": true}}, tenv{"WSL_INTEROP": "1"})
	if !s.IsMounted("/mnt/c") { t.Fatalf("expected mount true") }
	if !s.CanEditHostFiles() { t.Fatalf("expected can edit host files true") }
}
