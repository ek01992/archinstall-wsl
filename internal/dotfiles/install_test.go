package dotfiles

import (
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Existing legacy tests remain; add DI service tests below

type dfs struct {
	home     string
	files    map[string][]byte
	links    map[string]string
	entries  map[string][]string
	lstat    map[string]fs.FileMode
	readlink map[string]string
}

func (d *dfs) UserHomeDir() (string, error) { return d.home, nil }
func (d *dfs) WriteFile(path string, data []byte, perm fs.FileMode) error {
	if d.files == nil {
		d.files = map[string][]byte{}
	}
	d.files[path] = append([]byte(nil), data...)
	return nil
}
func (d *dfs) ReadDir(path string) ([]fs.DirEntry, error) {
	var out []fs.DirEntry
	for _, name := range d.entries[path] {
		n := name
		out = append(out, fakeDir{name: n})
	}
	return out, nil
}
func (d *dfs) Lstat(path string) (fs.FileInfo, error) {
	if m, ok := d.lstat[path]; ok {
		return fakeInfo{mode: m}, nil
	}
	return nil, fs.ErrNotExist
}
func (d *dfs) Readlink(path string) (string, error) {
	if s, ok := d.readlink[path]; ok {
		return s, nil
	}
	return "", fs.ErrNotExist
}
func (d *dfs) Symlink(oldname, newname string) error {
	if d.links == nil {
		d.links = map[string]string{}
	}
	d.links[newname] = oldname
	return nil
}

type fakeDir struct{ name string }

func (f fakeDir) Name() string               { return f.name }
func (f fakeDir) IsDir() bool                { return false }
func (f fakeDir) Type() fs.FileMode          { return 0 }
func (f fakeDir) Info() (fs.FileInfo, error) { return fakeInfo{}, nil }

type fakeInfo struct{ mode fs.FileMode }

func (f fakeInfo) Name() string       { return "" }
func (f fakeInfo) Size() int64        { return 0 }
func (f fakeInfo) Mode() fs.FileMode  { return f.mode }
func (f fakeInfo) ModTime() time.Time { return time.Unix(0, 0) }
func (f fakeInfo) IsDir() bool        { return f.mode.IsDir() }
func (f fakeInfo) Sys() any           { return nil }

type drun struct{ called []string }

func (d *drun) Run(name string, args ...string) error {
	d.called = append(d.called, name+" "+strings.Join(args, " "))
	return nil
}

func TestService_Install_DefaultZshrc(t *testing.T) {
	df := &dfs{home: "/home/alice"}
	s := NewService(df, &drun{})
	if err := s.Install("   "); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	z := df.files["/home/alice/.zshrc"]
	if len(z) == 0 || !strings.Contains(string(z), "oh-my-zsh") {
		t.Fatalf("expected default .zshrc written")
	}
}

func TestService_Install_CloneAndSymlink(t *testing.T) {
	df := &dfs{home: "/home/bob", entries: map[string][]string{"/home/bob/.dotfiles": {"zshrc", ".gitconfig", "README.md"}}, lstat: map[string]fs.FileMode{}}
	dr := &drun{}
	s := NewService(df, dr)
	if err := s.Install("https://example.com/dotfiles.git"); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	// git clone called
	if len(dr.called) == 0 || !strings.HasPrefix(dr.called[0], "git clone --depth 1 https://example.com/dotfiles.git ") {
		t.Fatalf("expected git clone, got %v", dr.called)
	}
	// symlinks
	if df.links[filepath.Join("/home/bob", ".zshrc")] != filepath.Join("/home/bob/.dotfiles", "zshrc") {
		t.Fatalf("zshrc link wrong: %v", df.links)
	}
	if df.links[filepath.Join("/home/bob", ".gitconfig")] != filepath.Join("/home/bob/.dotfiles", ".gitconfig") {
		t.Fatalf("gitconfig link wrong: %v", df.links)
	}
	if _, ok := df.links[filepath.Join("/home/bob", "README.md")]; ok {
		t.Fatalf("README should be skipped")
	}
}

func TestService_Install_IdempotentSkipWhenLinkMatches(t *testing.T) {
	df := &dfs{home: "/home/carol", entries: map[string][]string{"/home/carol/.dotfiles": {"zshrc"}}, lstat: map[string]fs.FileMode{filepath.Join("/home/carol", ".zshrc"): fs.ModeSymlink}, readlink: map[string]string{filepath.Join("/home/carol", ".zshrc"): filepath.Join("/home/carol/.dotfiles", "zshrc")}}
	s := NewService(df, &drun{})
	if err := s.Install("https://example.com/dotfiles.git"); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	// No new link should be created
	if len(df.links) != 0 {
		t.Fatalf("expected no new links, got %v", df.links)
	}
}
