package platform

import runtimepkg "archwsl-tui-configurator/internal/runtime"

func init() {
	fs := runtimepkg.NewFS()
	env := runtimepkg.NewEnv()
	readFile = func(path string) ([]byte, error) { return fs.ReadFile(path) }
	pathExists = func(path string) bool { _, err := fs.Stat(path); return err == nil }
	getenv = func(k string) string { return env.Getenv(k) }
}
