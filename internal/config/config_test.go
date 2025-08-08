package config

import (
	"io/fs"
	"reflect"
	"strings"
	"testing"
)

func TestSaveConfig_WritesYamlAtDefaultPath(t *testing.T) {
	origHome := getUserHomeDir
	origMkdir := mkdirAll
	origWrite := writeFile

	t.Cleanup(func() {
		getUserHomeDir = origHome
		mkdirAll = origMkdir
		writeFile = origWrite
	})

	getUserHomeDir = func() (string, error) { return "/home/alice", nil }

	var madeDir string
	var madePerm fs.FileMode
	mkdirAll = func(path string, perm fs.FileMode) error { madeDir, madePerm = path, perm; return nil }

	var wrotePath string
	var wrotePerm fs.FileMode
	var wroteData []byte
	writeFile = func(path string, data []byte, perm fs.FileMode) error {
		wrotePath = path
		wrotePerm = perm
		wroteData = data
		return nil
	}

	cfg := Config{
		Username:       "alice",
		GitName:        "Alice",
		GitEmail:       "alice@example.com",
		OhMyZshTheme:   "agnoster",
		OhMyZshPlugins: []string{"git", "fzf"},
		DotfilesRepo:   "https://example.com/dotfiles.git",
		NonInteractive: true,
	}

	if err := saveConfig(cfg); err != nil {
		t.Fatalf("saveConfig returned error: %v", err)
	}

	if madeDir != "/home/alice/.config/archwsl-tui-configurator" || madePerm != 0o700 {
		t.Fatalf("expected mkdirAll(0700) for config dir; got %q %v", madeDir, madePerm)
	}
	if wrotePath != "/home/alice/.config/archwsl-tui-configurator/config.yaml" || wrotePerm != 0o600 {
		t.Fatalf("unexpected config write path/perm: %q %v", wrotePath, wrotePerm)
	}
	if !strings.Contains(string(wroteData), "NonInteractive: true") {
		t.Fatalf("yaml should contain NonInteractive: true; got: %s", string(wroteData))
	}
}

func TestLoadConfig_ReadsYamlAndIsNonInteractive(t *testing.T) {
	origHome := getUserHomeDir
	origRead := readFile

	t.Cleanup(func() {
		getUserHomeDir = origHome
		readFile = origRead
	})

	getUserHomeDir = func() (string, error) { return "/home/bob", nil }

	yaml := "" +
		"Username: bob\n" +
		"GitName: Bob\n" +
		"GitEmail: bob@example.com\n" +
		"OhMyZshTheme: robbyrussell\n" +
		"OhMyZshPlugins:\n  - git\n  - z\n" +
		"DotfilesRepo: https://example.com/dots.git\n" +
		"NonInteractive: true\n"

	readFile = func(path string) ([]byte, error) {
		// Should use provided path
		if path != "/tmp/test-config.yaml" {
			t.Fatalf("unexpected load path: %q", path)
		}
		return []byte(yaml), nil
	}

	cfg := loadConfig("/tmp/test-config.yaml")
	if !cfg.NonInteractive {
		t.Fatalf("expected NonInteractive=true after load")
	}
	if cfg.Username != "bob" || cfg.GitEmail != "bob@example.com" {
		t.Fatalf("unexpected fields after load: %+v", cfg)
	}
}

func TestLoadConfig_EmptyPathUsesDefault(t *testing.T) {
	origHome := getUserHomeDir
	origRead := readFile

	t.Cleanup(func() { getUserHomeDir = origHome; readFile = origRead })

	getUserHomeDir = func() (string, error) { return "/home/carl", nil }

	readFile = func(path string) ([]byte, error) {
		if path != "/home/carl/.config/archwsl-tui-configurator/config.yaml" {
			t.Fatalf("expected default config path; got %q", path)
		}
		return []byte("NonInteractive: true\n"), nil
	}

	cfg := loadConfig("  \t\n")
	if !cfg.NonInteractive {
		t.Fatalf("expected NonInteractive=true after load from default path")
	}
}

func TestSaveAndLoad_PreservesConfig(t *testing.T) {
	// Full round-trip with seams capturing write then feeding to load
	origHome := getUserHomeDir
	origMkdir := mkdirAll
	origWrite := writeFile
	origRead := readFile

	t.Cleanup(func() { getUserHomeDir = origHome; mkdirAll = origMkdir; writeFile = origWrite; readFile = origRead })

	getUserHomeDir = func() (string, error) { return "/home/dana", nil }

	var bufPath string
	var buf []byte
	mkdirAll = func(path string, perm fs.FileMode) error { return nil }
	writeFile = func(path string, data []byte, perm fs.FileMode) error {
		bufPath = path
		buf = append([]byte(nil), data...)
		return nil
	}
	readFile = func(path string) ([]byte, error) {
		if path != bufPath {
			t.Fatalf("read path mismatch; got %q want %q", path, bufPath)
		}
		return buf, nil
	}

	want := Config{Username: "dana", GitName: "Dana", GitEmail: "dana@example.com", OhMyZshTheme: "agnoster", OhMyZshPlugins: []string{"git"}, NonInteractive: true}
	if err := saveConfig(want); err != nil {
		t.Fatalf("saveConfig error: %v", err)
	}
	got := loadConfig("")
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("round-trip mismatch:\nwant=%#v\n got=%#v", want, got)
	}
}
