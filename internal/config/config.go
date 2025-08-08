package config

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents persisted run configuration
// YAML field names are capitalized to match expected keys in tests
type Config struct {
	Username       string   `yaml:"Username"`
	GitName        string   `yaml:"GitName"`
	GitEmail       string   `yaml:"GitEmail"`
	OhMyZshTheme   string   `yaml:"OhMyZshTheme"`
	OhMyZshPlugins []string `yaml:"OhMyZshPlugins"`
	DotfilesRepo   string   `yaml:"DotfilesRepo"`
	NonInteractive bool     `yaml:"NonInteractive"`
}

var (
	getUserHomeDir = func() (string, error) { return osUserHomeDirImpl() }
	mkdirAll       = func(path string, perm fs.FileMode) error { return os.MkdirAll(path, perm) }
	writeFile      = func(path string, data []byte, perm fs.FileMode) error { return os.WriteFile(path, data, perm) }
	readFile       = func(path string) ([]byte, error) { return os.ReadFile(path) }
)

// saveConfig writes configuration YAML to ~/.config/archwsl-tui-configurator/config.yaml with secure perms
func saveConfig(cfg Config) error {
	home, err := getUserHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		return fmt.Errorf("cannot determine home directory: %v", err)
	}
	dir := filepath.Join(home, ".config", "archwsl-tui-configurator")
	if err := mkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("mkdir config dir: %w", err)
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal yaml: %w", err)
	}
	path := filepath.Join(dir, "config.yaml")
	if err := writeFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

// loadConfig reads configuration from the provided path; if empty/whitespace, uses default path.
// On read error, returns zero-value Config.
func loadConfig(path string) Config {
	path = strings.TrimSpace(path)
	if path == "" {
		home, err := getUserHomeDir()
		if err != nil || strings.TrimSpace(home) == "" {
			return Config{}
		}
		path = filepath.Join(home, ".config", "archwsl-tui-configurator", "config.yaml")
	}
	data, err := readFile(path)
	if err != nil {
		return Config{}
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}
	}
	return cfg
}

// osUserHomeDirImpl is separated for testing seams.
func osUserHomeDirImpl() (string, error) { return os.UserHomeDir() }
