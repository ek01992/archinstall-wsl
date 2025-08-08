package modules

import (
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

// Module represents a runnable module with name, optional description, and commands.
type Module struct {
	Name        string   `yaml:"name" toml:"name"`
	Description string   `yaml:"description" toml:"description"`
	Commands    []string `yaml:"commands" toml:"commands"`
}

func normalizeModule(m *Module) {
	m.Name = strings.TrimSpace(m.Name)
	m.Description = strings.TrimSpace(m.Description)
	for i := range m.Commands {
		m.Commands[i] = strings.TrimSpace(m.Commands[i])
	}
}

func validateModule(m Module) error {
	if m.Name == "" {
		return fmt.Errorf("module name must not be empty")
	}
	if len(m.Commands) == 0 {
		return fmt.Errorf("module %q must have at least one command", m.Name)
	}
	for i, c := range m.Commands {
		if strings.TrimSpace(c) == "" {
			return fmt.Errorf("module %q has empty command at index %d", m.Name, i)
		}
	}
	return nil
}

// LoadModulesYAML parses a YAML list of modules, normalizes fields, and validates them.
func LoadModulesYAML(data []byte) ([]Module, error) {
	var mods []Module
	if err := yaml.Unmarshal(data, &mods); err != nil {
		return nil, err
	}
	if len(mods) == 0 {
		return nil, fmt.Errorf("no modules defined")
	}
	for i := range mods {
		normalizeModule(&mods[i])
		if err := validateModule(mods[i]); err != nil {
			return nil, err
		}
	}
	return mods, nil
}

// LoadModulesTOML parses a TOML document containing [[modules]] tables.
func LoadModulesTOML(data []byte) ([]Module, error) {
	var doc struct {
		Modules []Module `toml:"modules"`
	}
	if err := toml.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	if len(doc.Modules) == 0 {
		return nil, fmt.Errorf("no modules defined")
	}
	for i := range doc.Modules {
		normalizeModule(&doc.Modules[i])
		if err := validateModule(doc.Modules[i]); err != nil {
			return nil, err
		}
	}
	return doc.Modules, nil
}
