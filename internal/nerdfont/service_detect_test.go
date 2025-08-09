package nerdfont

import "testing"

type fp struct{ wsl bool }

func (p fp) IsWSL() bool { return p.wsl }

type ffs struct {
	dirs  map[string][]string
	files map[string][]byte
}

func (f ffs) ReadDir(dir string) ([]string, error) {
	if f.dirs == nil {
		return nil, fsErrNotExist{}
	}
	if v, ok := f.dirs[dir]; ok {
		return v, nil
	}
	return nil, fsErrNotExist{}
}

func (f ffs) ReadFile(path string) ([]byte, error) {
	if f.files == nil {
		return nil, fsErrNotExist{}
	}
	if v, ok := f.files[path]; ok {
		return v, nil
	}
	return nil, fsErrNotExist{}
}

type fr struct {
	psOut   string
	wslPath string
}

func (r fr) PowerShell(args ...string) (string, error) { return r.psOut, nil }
func (r fr) WSLPath(args ...string) (string, error)    { return r.wslPath, nil }

type fsErrNotExist struct{}

func (fsErrNotExist) Error() string { return "not exist" }

func TestDetect_NonWSL_False(t *testing.T) {
	s := NewService(fp{wsl: false}, ffs{}, fr{})
	if s.Detect() {
		t.Fatalf("expected false when not WSL")
	}
}

func TestDetect_WSLShellPathHit_True(t *testing.T) {
	fs := ffs{dirs: map[string][]string{
		"/fonts": {"JetBrainsMono Nerd Font.ttf"},
	}}
	s := NewService(fp{wsl: true}, fs, fr{wslPath: "/fonts"})
	if !s.Detect() {
		t.Fatalf("expected true when WSLPath points to dir with Nerd Font")
	}
}
