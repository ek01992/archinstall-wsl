//go:build !windows

package runtime_test

import (
	"io/fs"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	rtx "archwsl-tui-configurator/internal/runtime"
)

func TestProdRunner_OutputEcho(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on windows")
	}
	runner := rtx.NewRunner(5 * time.Second)
	out, err := runner.Output("echo", "ok")
	if err != nil {
		t.Fatalf("Output returned error: %v", err)
	}
	if out != "ok\n" {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestProdRunner_RunTrue(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on windows")
	}
	runner := rtx.NewRunner(5 * time.Second)
	if err := runner.Run("sh", "-c", "true"); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
}

func TestProdFS_BasicOperations(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on windows")
	}
	fsys := rtx.NewFS()
	tmp := t.TempDir()

	file := filepath.Join(tmp, "file.txt")
	data := []byte("hello")
	if err := fsys.WriteFile(file, data, 0o644); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}
	read, err := fsys.ReadFile(file)
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}
	if string(read) != string(data) {
		t.Fatalf("ReadFile content mismatch: %q != %q", string(read), string(data))
	}

	// MkdirAll + ReadDir
	dir := filepath.Join(tmp, "dir", "sub")
	if err := fsys.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll error: %v", err)
	}
	entries, err := fsys.ReadDir(filepath.Join(tmp, "dir"))
	if err != nil {
		t.Fatalf("ReadDir error: %v", err)
	}
	if len(entries) != 1 || entries[0].Name() != "sub" {
		t.Fatalf("ReadDir unexpected entries: %+v", entries)
	}

	// Chmod + Lstat + Stat
	if err := fsys.Chmod(file, 0o600); err != nil {
		t.Fatalf("Chmod error: %v", err)
	}
	st1, err := fsys.Lstat(file)
	if err != nil {
		t.Fatalf("Lstat error: %v", err)
	}
	if st1.Mode().Perm() != fs.FileMode(0o600) {
		t.Fatalf("Lstat perms = %v, want 0600", st1.Mode().Perm())
	}
	st2, err := fsys.Stat(file)
	if err != nil {
		t.Fatalf("Stat error: %v", err)
	}
	if st2.Size() <= 0 {
		t.Fatalf("Stat size = %d, want > 0", st2.Size())
	}

	// Symlink + Readlink + Remove
	link := filepath.Join(tmp, "link")
	if err := fsys.Symlink(file, link); err != nil {
		t.Fatalf("Symlink error: %v", err)
	}
	target, err := fsys.Readlink(link)
	if err != nil {
		t.Fatalf("Readlink error: %v", err)
	}
	if target != file {
		t.Fatalf("Readlink target = %q, want %q", target, file)
	}
	if err := fsys.Remove(link); err != nil {
		t.Fatalf("Remove link error: %v", err)
	}
	if err := fsys.Remove(file); err != nil {
		t.Fatalf("Remove file error: %v", err)
	}
}

func TestProdEnv_Getenv(t *testing.T) {
	env := rtx.NewEnv()
	const key = "ARCHWSL_RUNTIME_TEST_KEY"
	t.Setenv(key, "value")
	if got := env.Getenv(key); got != "value" {
		t.Fatalf("Getenv(%q) = %q, want %q", key, got, "value")
	}
}
