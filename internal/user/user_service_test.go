package user

import (
	"errors"
	"io/fs"
	"strings"
	"testing"
)

type zfakeCmd struct{ calls []string; fail map[string]error; onRun func(name string, args ...string) }
func (c *zfakeCmd) Run(name string, args ...string) error {
	key := name + " " + strings.Join(args, " ")
	c.calls = append(c.calls, key)
	if c.onRun != nil { c.onRun(name, args...) }
	if c.fail != nil { if err, ok := c.fail[key]; ok { return err } }
	return nil
}
func (c *zfakeCmd) RunWithStdin(name, stdin string, args ...string) error { return c.Run(name, args...) }

type zfakeFS struct{ files map[string][]byte }
func (f *zfakeFS) ReadFile(path string) ([]byte, error) { if b, ok := f.files[path]; ok { return append([]byte(nil), b...), nil }; return nil, fs.ErrNotExist }
func (f *zfakeFS) WriteFile(path string, data []byte, perm fs.FileMode) error { if f.files==nil { f.files = map[string][]byte{} }; f.files[path] = append([]byte(nil), data...); return nil }
func (f *zfakeFS) MkdirAll(path string, perm fs.FileMode) error { return nil }
func (f *zfakeFS) Chmod(path string, mode fs.FileMode) error { return nil }

type zfakeLookup struct{ cur string; homes map[string]string; exists map[string]bool }
func (l *zfakeLookup) UserExists(username string) bool { return l.exists != nil && l.exists[username] }
func (l *zfakeLookup) HomeDirByUsername(username string) (string, error) { if h, ok := l.homes[username]; ok { return h, nil }; return "", errors.New("no home") }
func (l *zfakeLookup) CurrentUsername() string { return l.cur }

type zfakeSudo struct{}
func (zfakeSudo) Validate(content string) error { return nil }

func TestService_InstallZsh_UsesChshAndVerifies(t *testing.T) {
	cmd := &zfakeCmd{}
	fs := &zfakeFS{files: map[string][]byte{"/etc/passwd": []byte("alice:x:1000:1000::/home/alice:/bin/bash\n")}}
	lk := &zfakeLookup{cur: "alice"}
	s := NewService(cmd, fs, lk, zfakeSudo{})
	cmd.onRun = func(name string, args ...string) {
		if name == "chsh" && len(args) == 3 && args[0] == "-s" && args[1] == "/usr/bin/zsh" && args[2] == "alice" {
			fs.files["/etc/passwd"] = []byte("alice:x:1000:1000::/home/alice:/usr/bin/zsh\n")
		}
	}
	if err := s.InstallZsh(); err != nil { t.Fatalf("unexpected error: %v", err) }
}

func TestService_InstallZsh_FallbackToUsermod(t *testing.T) {
	cmd := &zfakeCmd{fail: map[string]error{"chsh -s /usr/bin/zsh bob": errors.New("no chsh")}}
	fs := &zfakeFS{files: map[string][]byte{"/etc/passwd": []byte("bob:x:1001:1001::/home/bob:/bin/bash\n")}}
	lk := &zfakeLookup{cur: "bob"}
	s := NewService(cmd, fs, lk, zfakeSudo{})
	cmd.onRun = func(name string, args ...string) {
		if name == "usermod" && len(args) == 3 && args[0] == "-s" && args[1] == "/usr/bin/zsh" && args[2] == "bob" {
			fs.files["/etc/passwd"] = []byte("bob:x:1001:1001::/home/bob:/usr/bin/zsh\n")
		}
	}
	if err := s.InstallZsh(); err != nil { t.Fatalf("unexpected error: %v", err) }
}

func TestService_InstallZsh_Idempotent(t *testing.T) {
	cmd := &zfakeCmd{}
	fs := &zfakeFS{files: map[string][]byte{"/etc/passwd": []byte("carol:x:1002:1002::/home/carol:/usr/bin/zsh\n")}}
	lk := &zfakeLookup{cur: "carol"}
	s := NewService(cmd, fs, lk, zfakeSudo{})
	if err := s.InstallZsh(); err != nil { t.Fatalf("unexpected error: %v", err) }
	for _, c := range cmd.calls { if strings.Contains(c, "chsh") || strings.Contains(c, "usermod") { t.Fatalf("no shell change expected") } }
}

func TestService_InstallZshTx_RollbackOnFailure(t *testing.T) {
	cmd := &zfakeCmd{fail: map[string]error{"chsh -s /usr/bin/zsh alice": errors.New("fail")}}
	fs := &zfakeFS{files: map[string][]byte{"/etc/passwd": []byte("alice:x:1000:1000::/home/alice:/bin/bash\n")}}
	lk := &zfakeLookup{cur: "alice"}
	s := NewService(cmd, fs, lk, zfakeSudo{})
	rolledBack := false
	cmd.onRun = func(name string, args ...string) {
		if name == "chsh" && len(args) == 3 && args[0] == "-s" && args[1] == "/bin/bash" && args[2] == "alice" { rolledBack = true }
	}
	if err := s.InstallZshTx(); err == nil { t.Fatalf("expected error") }
	if !rolledBack { t.Fatalf("expected rollback chsh to previous shell") }
}

func TestService_InstallOhMyZsh_WriteAndVerify(t *testing.T) {
	cmd := &zfakeCmd{}
	fs := &zfakeFS{files: map[string][]byte{}}
	lk := &zfakeLookup{homes: map[string]string{"dave": "/home/dave"}}
	s := NewService(cmd, fs, lk, zfakeSudo{})
	if err := s.InstallOhMyZsh("dave", "agnoster", []string{"git", "fzf"}); err != nil { t.Fatalf("unexpected: %v", err) }
	z := fs.files["/home/dave/.zshrc"]
	if len(z) == 0 || !strings.Contains(string(z), "ZSH_THEME=\"agnoster\"") { t.Fatalf(".zshrc not written as expected: %q", string(z)) }
}
