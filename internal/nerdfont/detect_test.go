package nerdfont

import "testing"

func TestDetectNerdFontInstalled_Positive(t *testing.T) {
	origEnum := enumerateWindowsFontFiles
	t.Cleanup(func() { enumerateWindowsFontFiles = origEnum })

	enumerateWindowsFontFiles = func() ([]string, error) {
		return []string{
			"Arial.ttf",
			"JetBrainsMono Nerd Font.ttf",
			"SomeOther.otf",
		}, nil
	}

	if !detectNerdFontInstalled() {
		t.Fatalf("expected true when Nerd Font present")
	}
}

func TestDetectNerdFontInstalled_Negative_NoNerdFonts(t *testing.T) {
	origEnum := enumerateWindowsFontFiles
	t.Cleanup(func() { enumerateWindowsFontFiles = origEnum })

	enumerateWindowsFontFiles = func() ([]string, error) {
		return []string{"Arial.ttf", "Calibri.otf", "FiraCode-Regular.ttf"}, nil
	}

	if detectNerdFontInstalled() {
		t.Fatalf("expected false when no Nerd Font present")
	}
}

func TestDetectNerdFontInstalled_HandlesErrorAsFalse(t *testing.T) {
	origEnum := enumerateWindowsFontFiles
	t.Cleanup(func() { enumerateWindowsFontFiles = origEnum })

	enumerateWindowsFontFiles = func() ([]string, error) { return nil, assertErr }

	if detectNerdFontInstalled() {
		t.Fatalf("expected false when enumeration fails")
	}
}

// assertErr is a sentinel error for tests; using a simple type to avoid importing errors
type errString string

func (e errString) Error() string { return string(e) }

const assertErr = errString("boom")
