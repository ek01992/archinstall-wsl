package user

import "io/fs"

type CommandRunner interface {
	Run(name string, args ...string) error
	RunWithStdin(name, stdin string, args ...string) error
}

type FS interface {
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte, perm fs.FileMode) error
	MkdirAll(path string, perm fs.FileMode) error
	Chmod(path string, mode fs.FileMode) error
}

type Lookup interface {
	UserExists(username string) bool
	HomeDirByUsername(username string) (string, error)
	CurrentUsername() string
}

type SudoersValidator interface {
	Validate(content string) error
}

type Service struct {
	cmd        CommandRunner
	fs         FS
	id         Lookup
	sudo       SudoersValidator
	zshPath    string
	sudoersDir string
}

func NewService(cmd CommandRunner, fs FS, id Lookup, sudo SudoersValidator) *Service {
	return &Service{cmd: cmd, fs: fs, id: id, sudo: sudo, zshPath: "/usr/bin/zsh", sudoersDir: ""}
}
