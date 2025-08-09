package dotfiles

import (
	"io/fs"
	runtimepkg "archwsl-tui-configurator/internal/runtime"
)

// Deprecated: use dotfiles.Service.Install. This shim delegates to a DI service.
func installDotfiles(repoURL string) error {
	fs := runtimepkg.NewFS()
	r := runtimepkg.NewRunner(30 * 1e9)
	return NewService(fsAdapter{fs: fs}, runAdapter{r: r}).Install(repoURL)
}

type fsAdapter struct{ fs runtimepkg.FS }
func (a fsAdapter) UserHomeDir() (string, error) { return a.fs.UserHomeDir() }
func (a fsAdapter) WriteFile(p string, d []byte, perm fs.FileMode) error { return a.fs.WriteFile(p, d, perm) }
func (a fsAdapter) ReadDir(p string) ([]fs.DirEntry, error) { return a.fs.ReadDir(p) }
func (a fsAdapter) Lstat(p string) (fs.FileInfo, error) { return a.fs.Lstat(p) }
func (a fsAdapter) Readlink(p string) (string, error) { return a.fs.Readlink(p) }
func (a fsAdapter) Symlink(old, new string) error { return a.fs.Symlink(old, new) }

type runAdapter struct{ r runtimepkg.Runner }
func (a runAdapter) Run(name string, args ...string) error { return a.r.Run(name, args...) }
