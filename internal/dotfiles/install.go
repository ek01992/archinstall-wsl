package dotfiles

import (
	"io/fs"
	"os"
	"os/exec"

	runtimepkg "archwsl-tui-configurator/internal/runtime"
)

// NOTE: Package-level seams below are for testability and are NOT concurrency-safe.
// These are deprecated in favor of dotfiles.Service. They remain temporarily for legacy tests.
var (
	getUserHomeDir = func() (string, error) { return osUserHomeDir() }
	osUserHomeDir  = func() (string, error) { return osUserHomeDirImpl() }

	// File/FS seams
	pathExists = func(path string) bool {
		_, err := os.Stat(path)
		return err == nil
	}
	writeFile = func(path string, data []byte, perm fs.FileMode) error { return os.WriteFile(path, data, perm) }
	listFiles = func(dir string) ([]string, error) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return nil, err
		}
		names := make([]string, 0, len(entries))
		for _, e := range entries {
			names = append(names, e.Name())
		}
		return names, nil
	}
	lstat = func(path string) (fs.FileMode, error) {
		fi, err := os.Lstat(path)
		if err != nil {
			return 0, err
		}
		return fi.Mode(), nil
	}
	readlink = func(path string) (string, error) { return os.Readlink(path) }
	symlink  = func(oldname string, newname string) error { return os.Symlink(oldname, newname) }

	// Command seam
	runCommand = func(name string, args ...string) error { return exec.Command(name, args...).Run() }
)

// osUserHomeDirImpl is split so tests can override getUserHomeDir only.
func osUserHomeDirImpl() (string, error) { return os.UserHomeDir() }

// installDotfiles is deprecated; use Service.Install with DI. This delegates to a Service wired with runtime deps.
func installDotfiles(repoURL string) error {
	fs := runtimepkg.NewFS()
	r := runtimepkg.NewRunner(30 * 1e9) // 30s
	return NewService(sFSAdapter{fs: fs}, runnerAdapter{r: r}).Install(repoURL)
}

type sFSAdapter struct{ fs runtimepkg.FS }

func (a sFSAdapter) UserHomeDir() (string, error) { return a.fs.UserHomeDir() }
func (a sFSAdapter) WriteFile(p string, d []byte, perm fs.FileMode) error {
	return a.fs.WriteFile(p, d, perm)
}
func (a sFSAdapter) ReadDir(p string) ([]fs.DirEntry, error) { return a.fs.ReadDir(p) }
func (a sFSAdapter) Lstat(p string) (fs.FileInfo, error)     { return a.fs.Lstat(p) }
func (a sFSAdapter) Readlink(p string) (string, error)       { return a.fs.Readlink(p) }
func (a sFSAdapter) Symlink(old, new string) error           { return a.fs.Symlink(old, new) }

type runnerAdapter struct{ r runtimepkg.Runner }

func (a runnerAdapter) Run(name string, args ...string) error { return a.r.Run(name, args...) }
