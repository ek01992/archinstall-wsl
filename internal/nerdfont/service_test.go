package nerdfont

import "testing"

type pf struct{}
func (pf) IsWSL() bool { return true }

type fsx struct{ names map[string][]string }
func (f fsx) ReadDir(dir string) ([]string, error) { if v, ok := f.names[dir]; ok { return v, nil }; return nil, assertErr }
func (fsx) ReadFile(path string) ([]byte, error) { return nil, nil }

type rf struct{}
func (rf) PowerShell(args ...string) (string, error) { return "", assertErr }
func (rf) WSLPath(args ...string) (string, error) { return "", assertErr }

func TestService_Detect_Positive(t *testing.T) {
	s := NewService(pf{}, fsx{names: map[string][]string{"/mnt/c/Windows/Fonts": {"JetBrainsMono Nerd Font.ttf"}}}, rf{})
	if !s.Detect() { t.Fatalf("expected true") }
}
