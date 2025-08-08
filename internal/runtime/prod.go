package runtime

import (
	"context"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// NewRunner returns a Runner that executes commands with a context timeout.
func NewRunner(timeout time.Duration) Runner {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &prodRunner{timeout: timeout}
}

type prodRunner struct {
	timeout time.Duration
}

func (p *prodRunner) Run(name string, args ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (p *prodRunner) Output(name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// NewFS returns an FS backed by the local OS filesystem.
func NewFS() FS { return prodFS{} }

type prodFS struct{}

func (prodFS) ReadFile(path string) ([]byte, error)                 { return os.ReadFile(path) }
func (prodFS) WriteFile(path string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(path, data, perm)
}
func (prodFS) MkdirAll(path string, perm fs.FileMode) error { return os.MkdirAll(path, perm) }
func (prodFS) Chmod(path string, mode fs.FileMode) error    { return os.Chmod(path, mode) }
func (prodFS) Lstat(path string) (fs.FileInfo, error)       { return os.Lstat(path) }
func (prodFS) ReadDir(path string) ([]fs.DirEntry, error)   { return os.ReadDir(path) }
func (prodFS) Stat(path string) (fs.FileInfo, error)        { return os.Stat(path) }
func (prodFS) Remove(path string) error                     { return os.Remove(path) }
func (prodFS) Readlink(path string) (string, error)         { return os.Readlink(path) }
func (prodFS) Symlink(oldname, newname string) error {
	// Ensure parent directory exists to behave nicely in tests
	if dir := filepath.Dir(newname); dir != "." {
		_ = os.MkdirAll(dir, 0o755)
	}
	return os.Symlink(oldname, newname)
}
func (prodFS) UserHomeDir() (string, error) { return os.UserHomeDir() }

// NewEnv returns an Env backed by os.Getenv.
func NewEnv() Env { return prodEnv{} }

type prodEnv struct{}

func (prodEnv) Getenv(key string) string { return os.Getenv(key) }
