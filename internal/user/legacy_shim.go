package user

import (
	"io/fs"
	"time"

	runtimepkg "archwsl-tui-configurator/internal/runtime"
)

// TODO: remove this legacy default service after DI migration completes.

// runtime-backed adapters

type runtimeCmd struct{ r runtimepkg.Runner }

func (rc runtimeCmd) Run(name string, args ...string) error { return rc.r.Run(name, args...) }
func (rc runtimeCmd) RunWithStdin(name, stdin string, args ...string) error {
	cmdline := name
	for _, a := range args { cmdline += " " + a }
	return rc.r.Run("sh", "-c", "printf %s \"$1\" | "+cmdline, "_", stdin)
}

type runtimeFS struct{ fs runtimepkg.FS }

func (rfs runtimeFS) ReadFile(path string) ([]byte, error) { return rfs.fs.ReadFile(path) }
func (rfs runtimeFS) WriteFile(path string, data []byte, perm fs.FileMode) error { return rfs.fs.WriteFile(path, data, perm) }
func (rfs runtimeFS) MkdirAll(path string, perm fs.FileMode) error { return rfs.fs.MkdirAll(path, perm) }
func (rfs runtimeFS) Chmod(path string, mode fs.FileMode) error { return rfs.fs.Chmod(path, mode) }

type runtimeLookup struct{}

func (runtimeLookup) UserExists(username string) bool { return doesUserExist(username) }
func (runtimeLookup) HomeDirByUsername(username string) (string, error) { return getHomeDirByUsername(username) }
func (runtimeLookup) CurrentUsername() string { return getTargetUsername() }

type runtimeSudo struct{}

func (runtimeSudo) Validate(content string) error { return validateSudoersContent(content) }

var defaultService = NewService(runtimeCmd{r: runtimepkg.NewRunner(10 * time.Second)}, runtimeFS{fs: runtimepkg.NewFS()}, runtimeLookup{}, runtimeSudo{})
