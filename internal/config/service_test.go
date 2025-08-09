package config

import (
	"io/fs"
	"reflect"
	"strings"
	"testing"
)

type sFakeFS struct {
	reads     map[string][]byte
	madePath  string
	madePerm  fs.FileMode
	wrotePath string
	wrotePerm fs.FileMode
	wroteData []byte
	readErr   error
}

func (f *sFakeFS) ReadFile(path string) ([]byte, error) {
	if f.readErr != nil {
		return nil, f.readErr
	}
	if f.reads == nil {
		return nil, fs.ErrNotExist
	}
	b, ok := f.reads[path]
	if !ok {
		return nil, fs.ErrNotExist
	}
	return b, nil
}

func (f *sFakeFS) WriteFile(path string, data []byte, perm fs.FileMode) error {
	f.wrotePath = path
	f.wrotePerm = perm
	f.wroteData = append([]byte(nil), data...)
	return nil
}

func (f *sFakeFS) MkdirAll(path string, perm fs.FileMode) error {
	f.madePath = path
	f.madePerm = perm
	return nil
}

func TestService_Save_WritesYamlAtDefaultPath(t *testing.T) {
	fsx := &sFakeFS{}
	home := func() (string, error) { return "/home/alice", nil }
	svc := NewService(fsx, home)

	cfg := Config{Username: "alice", NonInteractive: true}
	if err := svc.Save(cfg); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	if fsx.madePath != "/home/alice/.config/archwsl-tui-configurator" || fsx.madePerm != 0o700 {
		t.Fatalf("expected MkdirAll with 0700; got %q %v", fsx.madePath, fsx.madePerm)
	}
	if fsx.wrotePath != "/home/alice/.config/archwsl-tui-configurator/config.yaml" || fsx.wrotePerm != 0o600 {
		t.Fatalf("unexpected write path/perm: %q %v", fsx.wrotePath, fsx.wrotePerm)
	}
	if !strings.Contains(string(fsx.wroteData), "NonInteractive: true") {
		t.Fatalf("yaml should contain NonInteractive: true; got: %s", string(fsx.wroteData))
	}
}

func TestService_Load_ReadsYamlFromExplicitPath(t *testing.T) {
	yaml := "" +
		"Username: bob\n" +
		"GitEmail: bob@example.com\n" +
		"NonInteractive: true\n"
	fsx := &sFakeFS{reads: map[string][]byte{"/tmp/test-config.yaml": []byte(yaml)}}
	home := func() (string, error) { return "/ignored", nil }
	svc := NewService(fsx, home)

	cfg := svc.Load("/tmp/test-config.yaml")
	if !cfg.NonInteractive {
		t.Fatalf("expected NonInteractive=true after load")
	}
	if cfg.Username != "bob" || cfg.GitEmail != "bob@example.com" {
		t.Fatalf("unexpected fields after load: %+v", cfg)
	}
}

func TestService_Load_EmptyPathUsesDefault(t *testing.T) {
	fsx := &sFakeFS{reads: map[string][]byte{"/home/carl/.config/archwsl-tui-configurator/config.yaml": []byte("NonInteractive: true\n")}}
	home := func() (string, error) { return "/home/carl", nil }
	svc := NewService(fsx, home)

	cfg := svc.Load("  \t\n")
	if !cfg.NonInteractive {
		t.Fatalf("expected NonInteractive=true after load from default path")
	}
}

func TestService_Save_HomeError(t *testing.T) {
	fsx := &sFakeFS{}
	home := func() (string, error) { return "", fs.ErrPermission }
	svc := NewService(fsx, home)
	if err := svc.Save(Config{}); err == nil {
		t.Fatalf("expected error when home resolution fails")
	}
}

func TestService_Load_HomeErrorReturnsZeroValue(t *testing.T) {
	fsx := &sFakeFS{}
	home := func() (string, error) { return "", fs.ErrPermission }
	svc := NewService(fsx, home)
	got := svc.Load("\n\t ")
	want := Config{}
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("expected zero-value Config; got %#v", got)
	}
}
