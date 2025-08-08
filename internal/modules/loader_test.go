package modules

import (
	"testing"
)

func TestLoadModulesYAML_Valid(t *testing.T) {
	data := `
- name: " Go Toolchain "
  description: "Install Go"
  commands:
    - " pacman -S --noconfirm go "
    - "echo done"
- name: "Python"
  commands: ["pacman -S --noconfirm python"]
`

	mods, err := LoadModulesYAML([]byte(data))
	if err != nil {
		t.Fatalf("LoadModulesYAML returned error: %v", err)
	}
	if len(mods) != 2 {
		t.Fatalf("expected 2 modules, got %d", len(mods))
	}

	if mods[0].Name != "Go Toolchain" {
		t.Fatalf("expected trimmed name 'Go Toolchain', got %q", mods[0].Name)
	}
	if len(mods[0].Commands) != 2 || mods[0].Commands[0] != "pacman -S --noconfirm go" || mods[0].Commands[1] != "echo done" {
		t.Fatalf("unexpected commands for first module: %#v", mods[0].Commands)
	}
}

func TestLoadModulesYAML_Invalid(t *testing.T) {
	cases := []string{
		`not: [valid`,                         // malformed YAML
		`- name: ""\n  commands: ["echo hi"]`, // empty name
		`- name: X\n  commands: []`,           // empty commands
		`- name: X\n  commands: ["", "ok"]`,   // empty command entry
	}
	for i, data := range cases {
		if _, err := LoadModulesYAML([]byte(data)); err == nil {
			t.Fatalf("case %d: expected error for invalid YAML/modules", i)
		}
	}
}

func TestLoadModulesTOML_Valid(t *testing.T) {
	data := `
[[modules]]
name = "Node"
description = "Install Node"
commands = [" pacman -S --noconfirm nodejs ", "npm -v"]

[[modules]]
name = "Dotfiles"
commands = ["echo setup"]
`
	mods, err := LoadModulesTOML([]byte(data))
	if err != nil {
		t.Fatalf("LoadModulesTOML returned error: %v", err)
	}
	if len(mods) != 2 {
		t.Fatalf("expected 2 modules, got %d", len(mods))
	}
	if mods[0].Name != "Node" || mods[0].Commands[0] != "pacman -S --noconfirm nodejs" {
		t.Fatalf("unexpected normalization for first module: %+v", mods[0])
	}
}

func TestLoadModulesTOML_Invalid(t *testing.T) {
	cases := []string{
		`name = "X"`,
		`[[modules]]
name = ""
commands = ["echo"]
`,
		`[[modules]]
name = "X"
commands = []
`,
		`[[modules]]
name = "X"
commands = ["", "echo"]
`,
	}
	for i, data := range cases {
		if _, err := LoadModulesTOML([]byte(data)); err == nil {
			t.Fatalf("case %d: expected error for invalid TOML/modules", i)
		}
	}
}
