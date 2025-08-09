package nerdfont

import "testing"

func TestDetectNerdFontInstalled_UsesWSLConfRoot(t *testing.T) {
	// Covered by service-level test using DI; keep legacy test minimal
	orig := readFileConf
	origEnum := enumerateWindowsFontFiles
	origIsWSL := isWSL
	origReadDir := readFontDir
	t.Cleanup(func() {
		readFileConf = orig
		enumerateWindowsFontFiles = origEnum
		isWSL = origIsWSL
		readFontDir = origReadDir
	})
	isWSL = func() bool { return true }
	readFileConf = func(path string) ([]byte, error) { return []byte("[automount]\nroot = /altmnt/\n"), nil }
	readFontDir = func(dir string) ([]string, error) {
		if dir == "/altmnt/c/Windows/Fonts" {
			return []string{"JetBrainsMono Nerd Font.ttf"}, nil
		}
		return nil, assertErr
	}
	if !detectNerdFontInstalled() {
		t.Fatalf("expected true when Nerd Font present via wsl.conf root")
	}
}
