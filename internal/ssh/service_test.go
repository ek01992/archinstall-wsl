package ssh

import (
	"errors"
	"io/fs"
	"path/filepath"
	"testing"
)

type pf struct{ ok bool }
func (p pf) CanEditHostFiles() bool { return p.ok }

type fsfake struct{ files map[string][]byte; names []string; home string }
func (f *fsfake) ReadDir(dir string) ([]string, error) { if f.names!=nil { return append([]string(nil), f.names...), nil }; return nil, errors.New("no dir") }
func (f *fsfake) ReadFile(path string) ([]byte, error) { b, ok := f.files[path]; if !ok { return nil, errors.New("no file") }; return append([]byte(nil), b...), nil }
func (f *fsfake) WriteFile(path string, data []byte, perm fs.FileMode) error { if f.files==nil { f.files = map[string][]byte{} }; f.files[path] = append([]byte(nil), data...); return nil }
func (f *fsfake) MkdirAll(path string, perm fs.FileMode) error { return nil }
func (f *fsfake) Chmod(path string, mode fs.FileMode) error { return nil }
func (f *fsfake) UserHomeDir() (string, error) { return f.home, nil }

func TestService_ImportWithConsent_Succeeds(t *testing.T) {
	fs := &fsfake{home: "/home/alice", names: []string{"id_rsa"}, files: map[string][]byte{"/mnt/c/Users/Alice/.ssh/id_rsa": []byte("KEY")}}
	s := NewService(pf{ok:true}, fs)
	if err := s.ImportWithConsent("/mnt/c/Users/Alice/.ssh", true); err != nil { t.Fatalf("unexpected error: %v", err) }
	if _, ok := fs.files[filepath.Join("/home/alice/.ssh", "id_rsa")]; !ok { t.Fatalf("expected file copied") }
}
