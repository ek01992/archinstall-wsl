package nerdfont

import "testing"

func TestEnumerateFonts_AlternateMountRoot(t *testing.T) {
	origRead := readFontDir
	origIsWSL := isWSL
	t.Cleanup(func() { readFontDir = origRead; isWSL = origIsWSL })

	isWSL = func() bool { return true }

	readFontDir = func(dir string) ([]string, error) {
		if dir == "/c/Windows/Fonts" {
			return []string{"JetBrainsMono Nerd Font.ttf", "Arial.ttf"}, nil
		}
		return nil, assertErr
	}

	if !detectNerdFontInstalled() {
		t.Fatalf("expected true when Nerd Font present via alternate /c mount")
	}
}

func TestDetectNerdFontInstalled_RegistryFallback_Positive(t *testing.T) {
	origEnum := enumerateWindowsFontFiles
	origPS := runPSCapture
	origIsWSL := isWSL
	t.Cleanup(func() { enumerateWindowsFontFiles = origEnum; runPSCapture = origPS; isWSL = origIsWSL })

	// Force enumeration failure
	enumerateWindowsFontFiles = func() ([]string, error) { return nil, assertErr }
	// Pretend we're in WSL and registry contains a Nerd Font
	isWSL = func() bool { return true }
	runPSCapture = func(args ...string) (string, error) {
		return "... JetBrainsMono Nerd Font ...", nil
	}

	if !detectNerdFontInstalled() {
		t.Fatalf("expected true when registry shows Nerd Font")
	}
}
