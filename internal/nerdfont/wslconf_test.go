package nerdfont

import (
	"testing"
)

func TestParseWSLAutomountRoot_NoFile(t *testing.T) {
	orig := readFileConf
	t.Cleanup(func() { readFileConf = orig })
	readFileConf = func(path string) ([]byte, error) { return nil, assertErr }
	if got := parseWSLAutomountRoot(); got != "" {
		t.Fatalf("expected empty root when wsl.conf missing; got %q", got)
	}
}

func TestParseWSLAutomountRoot_ParsesRootValue(t *testing.T) {
	orig := readFileConf
	t.Cleanup(func() { readFileConf = orig })
	readFileConf = func(path string) ([]byte, error) {
		return []byte(`
# comment
[automount]
root = /altmnt/

[network]
hostname = arch
`), nil
	}
	if got := parseWSLAutomountRoot(); got != "/altmnt/" {
		t.Fatalf("unexpected root: %q", got)
	}
}

func TestParseWSLAutomountRoot_IgnoresOtherSectionsAndWhitespace(t *testing.T) {
	orig := readFileConf
	t.Cleanup(func() { readFileConf = orig })
	readFileConf = func(path string) ([]byte, error) {
		return []byte(`
; another comment
   [AUTOMOUNT]  
  root =   /mnt/   
`), nil
	}
	if got := parseWSLAutomountRoot(); got != "/mnt/" {
		t.Fatalf("expected trimmed /mnt/, got %q", got)
	}
}
