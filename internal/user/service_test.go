package user

import (
	"errors"
	"io/fs"
	"sync"
	"testing"
)

type fakeCmd struct{ mu sync.Mutex; runs []string }
func (f *fakeCmd) Run(name string, args ...string) error { f.mu.Lock(); defer f.mu.Unlock(); f.runs = append(f.runs, name); return nil }
func (f *fakeCmd) RunWithStdin(name, stdin string, args ...string) error { return f.Run(name, args...) }

type fakeFS struct{ mu sync.Mutex; files map[string][]byte }
func (f *fakeFS) ReadFile(path string) ([]byte, error) { f.mu.Lock(); defer f.mu.Unlock(); b, ok := f.files[path]; if !ok { return nil, errors.New("no file") }; return append([]byte(nil), b...), nil }
func (f *fakeFS) WriteFile(path string, data []byte, perm fs.FileMode) error { f.mu.Lock(); defer f.mu.Unlock(); if f.files==nil { f.files = map[string][]byte{} }; f.files[path] = append([]byte(nil), data...); return nil }
func (f *fakeFS) MkdirAll(path string, perm fs.FileMode) error { return nil }
func (f *fakeFS) Chmod(path string, mode fs.FileMode) error { return nil }

type fakeLookup struct{ exists map[string]bool }
func (l *fakeLookup) UserExists(username string) bool { return l.exists[username] }
func (l *fakeLookup) HomeDirByUsername(username string) (string, error) { return "/home/"+username, nil }
func (l *fakeLookup) CurrentUsername() string { return "root" }

type fakeSudo struct{}
func (fakeSudo) Validate(content string) error { return nil }

func TestUserService_Concurrent_NoRaces(t *testing.T) {
	cmd1, cmd2 := &fakeCmd{}, &fakeCmd{}
	fs1, fs2 := &fakeFS{}, &fakeFS{}
	lk1 := &fakeLookup{exists: map[string]bool{"alice": false}}
	lk2 := &fakeLookup{exists: map[string]bool{"bob": false}}
	s := NewService(cmd1, fs1, lk1, fakeSudo{})
	s2 := NewService(cmd2, fs2, lk2, fakeSudo{})

	var wg sync.WaitGroup
	wg.Add(2)
	go func(){ defer wg.Done(); _ = s.CreateUser("alice", "pw") }()
	go func(){ defer wg.Done(); _ = s2.CreateUser("bob", "pw") }()
	wg.Wait()
}
