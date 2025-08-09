package ssh

import (
	"errors"
	"io/fs"
	"sync"
	"testing"
)

type cfs struct {
	homes map[string]string
	files map[string][]byte
	dirs  map[string][]string
}

func (f *cfs) ReadDir(dir string) ([]string, error) {
	if v, ok := f.dirs[dir]; ok {
		return append([]string(nil), v...), nil
	}
	return nil, errors.New("no dir")
}
func (f *cfs) ReadFile(path string) ([]byte, error) {
	if b, ok := f.files[path]; ok {
		return append([]byte(nil), b...), nil
	}
	return nil, fs.ErrNotExist
}
func (f *cfs) WriteFile(path string, data []byte, perm fs.FileMode) error {
	if f.files == nil {
		f.files = map[string][]byte{}
	}
	f.files[path] = append([]byte(nil), data...)
	return nil
}
func (f *cfs) MkdirAll(path string, perm fs.FileMode) error { return nil }
func (f *cfs) Chmod(path string, mode fs.FileMode) error    { return nil }
func (f *cfs) UserHomeDir() (string, error)                 { return "/home/alice", nil }

type cplat struct{}

func (cplat) CanEditHostFiles() bool { return true }

func TestSSHService_Concurrent_NoRaces(t *testing.T) {
	fs1 := &cfs{dirs: map[string][]string{"/host1": {"id_rsa", "id_rsa.pub"}}, files: map[string][]byte{"/host1/id_rsa": []byte("k1"), "/host1/id_rsa.pub": []byte("k1pub")}}
	fs2 := &cfs{dirs: map[string][]string{"/host2": {"id_ecdsa"}}, files: map[string][]byte{"/host2/id_ecdsa": []byte("k2")}}
	s1 := NewService(cplat{}, fs1)
	s2 := NewService(cplat{}, fs2)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); _ = s1.ImportWithConsent("/host1", true) }()
	go func() { defer wg.Done(); _ = s2.ImportWithConsent("/host2", true) }()
	wg.Wait()
}
