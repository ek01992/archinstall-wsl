package nerdfont

import (
	"io/fs"
	"testing"
)

type pf struct{}

func (pf) IsWSL() bool { return true }

type fsx struct {
	names map[string][]string
	files map[string][]byte
}

func (f fsx) ReadDir(dir string) ([]string, error) {
	if v, ok := f.names[dir]; ok { return v, nil }
	return nil, fs.ErrNotExist
}
func (f fsx) ReadFile(path string) ([]byte, error) {
	if b, ok := f.files[path]; ok { return b, nil }
	return nil, fs.ErrNotExist
}

type rf struct{}

func (rf) PowerShell(args ...string) (string, error) { return "", fs.ErrNotExist }
func (rf) WSLPath(args ...string) (string, error)    { return "", fs.ErrNotExist }

func TestService_Detect_Positive(t *testing.T) {
	s := NewService(pf{}, fsx{names: map[string][]string{"/mnt/c/Windows/Fonts": {"JetBrainsMono Nerd Font.ttf"}}}, rf{})
	if !s.Detect() { t.Fatalf("expected true") }
}

func TestService_Detect_UsesWSLConfRoot(t *testing.T) {
	fs := fsx{
		names: map[string][]string{"/altmnt/c/Windows/Fonts": {"JetBrainsMono Nerd Font.ttf"}},
		files: map[string][]byte{"/etc/wsl.conf": []byte("[automount]\nroot = /altmnt/\n")},
	}
	s := NewService(pf{}, fs, rf{})
	if !s.Detect() { t.Fatalf("expected true when Nerd Font present via wsl.conf root") }
}
