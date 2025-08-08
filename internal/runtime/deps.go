package runtime

import "io/fs"

// Runner abstracts running external commands.
// Implementations should enforce sensible timeouts.
type Runner interface {
	Run(name string, args ...string) error
	Output(name string, args ...string) (string, error)
}

// FS abstracts filesystem operations used by the application.
// It enables testing and seam injection.
// Methods mirror common os/io functions.
type FS interface {
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte, perm fs.FileMode) error
	MkdirAll(path string, perm fs.FileMode) error
	Chmod(path string, mode fs.FileMode) error
	Lstat(path string) (fs.FileInfo, error)
	ReadDir(path string) ([]fs.DirEntry, error)
	Stat(path string) (fs.FileInfo, error)
	Remove(path string) error
	Readlink(path string) (string, error)
	Symlink(oldname, newname string) error
	UserHomeDir() (string, error)
}

// Env abstracts environment variable access.
// Implementations typically delegate to os.Getenv.
type Env interface {
	Getenv(key string) string
}

// Constructors for production implementations are provided in prod.go:
//   - NewRunner(timeout time.Duration) Runner
//   - NewFS() FS
//   - NewEnv() Env
