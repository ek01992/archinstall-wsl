package platform

import "testing"

func TestIsWSL_PositiveByEnv(t *testing.T) {
	orig := getenv
	defer func() { getenv = orig }()
	getenv = func(k string) string {
		if k == "WSL_INTEROP" { return "1" }
		return ""
	}
	if !IsWSL() { t.Fatalf("expected IsWSL true when env vars set") }
}

func TestIsWSL_PositiveByProc(t *testing.T) {
	orig := readFile
	defer func() { readFile = orig }()
	readFile = func(path string) ([]byte, error) {
		if path == "/proc/sys/kernel/osrelease" { return []byte("5.15.0-microsoft-standard-WSL2"), nil }
		return []byte(""), nil
	}
	if !IsWSL() { t.Fatalf("expected IsWSL true when osrelease mentions microsoft") }
}

func TestIsWSL_Negative(t *testing.T) {
	origEnv := getenv
	origRead := readFile
	defer func() { getenv = origEnv; readFile = origRead }()
	getenv = func(k string) string { return "" }
	readFile = func(path string) ([]byte, error) { return []byte("linux generic"), nil }
	if IsWSL() { t.Fatalf("expected IsWSL false on generic linux") }
}

func TestIsMounted(t *testing.T) {
	orig := pathExists
	defer func() { pathExists = orig }()
	pathExists = func(p string) bool { return p == "/mnt/c" }
	if !IsMounted("/mnt/c") || IsMounted("/nope") {
		t.Fatalf("unexpected IsMounted results")
	}
}

func TestCanEditHostFiles(t *testing.T) {
	origEnv := getenv
	origExists := pathExists
	defer func() { getenv = origEnv; pathExists = origExists }()
	getenv = func(k string) string { if k == "WSL_INTEROP" { return "1" }; return "" }
	pathExists = func(p string) bool { return p == "/mnt/c" }
	if !CanEditHostFiles() { t.Fatalf("expected true when WSL and /mnt/c present") }
}
