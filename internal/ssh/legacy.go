package ssh

import (
	"io/fs"

	"archwsl-tui-configurator/internal/platform"
	runtimepkg "archwsl-tui-configurator/internal/runtime"
)

// TODO: remove this legacy shim after DI migration completes.

// defaultService is a package-level Service built from production deps for callers that
// have not migrated to DI yet.
var defaultService = NewService(platformDefault{}, fsRuntimeAdapter{fs: runtimepkg.NewFS()})

type platformDefault struct{}

func (platformDefault) CanEditHostFiles() bool { return platform.CanEditHostFiles() }

type fsRuntimeAdapter struct{ fs runtimepkg.FS }

func (a fsRuntimeAdapter) ReadDir(dir string) ([]string, error) {
	entries, err := a.fs.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		// Only include files, mirroring legacy behavior
		if e.Type().IsDir() {
			continue
		}
		names = append(names, e.Name())
	}
	return names, nil
}
func (a fsRuntimeAdapter) ReadFile(path string) ([]byte, error) { return a.fs.ReadFile(path) }
func (a fsRuntimeAdapter) WriteFile(path string, data []byte, perm fs.FileMode) error {
	return a.fs.WriteFile(path, data, perm)
}
func (a fsRuntimeAdapter) MkdirAll(path string, perm fs.FileMode) error {
	return a.fs.MkdirAll(path, perm)
}
func (a fsRuntimeAdapter) Chmod(path string, mode fs.FileMode) error { return a.fs.Chmod(path, mode) }
func (a fsRuntimeAdapter) UserHomeDir() (string, error)              { return a.fs.UserHomeDir() }

// Deprecated: prefer constructing ssh.Service with DI. This shim will be removed.
func ImportFromWindows(hostPath string) error { return defaultService.ImportFromWindows(hostPath) }

// Deprecated: prefer constructing ssh.Service with DI. This shim will be removed.
func ImportWithConsent(hostPath string, consent bool) error {
	return defaultService.ImportWithConsent(hostPath, consent)
}
