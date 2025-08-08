package user

import (
	"io/fs"
	"time"
)

// TODO: remove this legacy default service after DI migration completes.

// seam-backed adapters use the package-level seams so legacy tests can override behavior.

type seamCmd struct{}

func (seamCmd) Run(name string, args ...string) error { return runCommand(name, args...) }
func (seamCmd) RunWithStdin(name, stdin string, args ...string) error { return runCommandWithStdin(name, stdin, args...) }

type seamFS struct{}

func (seamFS) ReadFile(path string) ([]byte, error) { return readFile(path) }
func (seamFS) WriteFile(path string, data []byte, perm fs.FileMode) error { return writeFile(path, data, perm) }
func (seamFS) MkdirAll(path string, perm fs.FileMode) error { return mkdirAll(path, perm) }
func (seamFS) Chmod(path string, mode fs.FileMode) error { return nil }

type seamLookup struct{}

func (seamLookup) UserExists(username string) bool { return doesUserExist(username) }
func (seamLookup) HomeDirByUsername(username string) (string, error) { return getHomeDirByUsername(username) }
func (seamLookup) CurrentUsername() string { return getTargetUsername() }

type seamSudo struct{}

func (seamSudo) Validate(content string) error { return validateSudoersContent(content) }

var defaultService = NewService(seamCmd{}, seamFS{}, seamLookup{}, seamSudo{})


// Keep a no-op variable reference to avoid unused import removal if needed
var _ = time.Second
