package app

import (
	"io/fs"
	"time"
	stduser "os/user"

	"archwsl-tui-configurator/internal/config"
	"archwsl-tui-configurator/internal/firewall"
	"archwsl-tui-configurator/internal/git"
	"archwsl-tui-configurator/internal/nerdfont"
	"archwsl-tui-configurator/internal/platform"
	runtimepkg "archwsl-tui-configurator/internal/runtime"
	"archwsl-tui-configurator/internal/ssh"
	"archwsl-tui-configurator/internal/toolchain/golang"
	"archwsl-tui-configurator/internal/toolchain/nodejs"
	"archwsl-tui-configurator/internal/toolchain/python"
	"archwsl-tui-configurator/internal/user"
)

// Provider builds default DI services backed by production runtime seams.
type Provider struct {
	User            *user.Service
	SSH             *ssh.Service
	Git             *git.Service
	Firewall        *firewall.Service
	NerdFont        *nerdfont.Service
	GoToolchain     *golang.Service
	NodeToolchain   *nodejs.Service
	PythonToolchain *python.Service
	Platform        *platform.Service
	Config          *config.Service
}

// Generic runner adapter for packages expecting Run/Output
type runnerAdapter struct{ r runtimepkg.Runner }
func (a runnerAdapter) Run(name string, args ...string) error            { return a.r.Run(name, args...) }
func (a runnerAdapter) Output(name string, args ...string) (string, error) { return a.r.Output(name, args...) }

// nodejs runner adapter adds Shell
type nodeRunnerAdapter struct{ r runtimepkg.Runner }
func (a nodeRunnerAdapter) Run(name string, args ...string) error            { return a.r.Run(name, args...) }
func (a nodeRunnerAdapter) Output(name string, args ...string) (string, error) { return a.r.Output(name, args...) }
func (a nodeRunnerAdapter) Shell(cmd string) (string, error)                 { return a.r.Output("bash", "-lc", cmd) }

// ssh adapters
type sshPlatformAdapter struct{ p *platform.Service }
func (a sshPlatformAdapter) CanEditHostFiles() bool { return a.p.CanEditHostFiles() }

type sshFSAdapter struct{ fs runtimepkg.FS }
func (a sshFSAdapter) ReadDir(dir string) ([]string, error) {
	entries, err := a.fs.ReadDir(dir)
	if err != nil { return nil, err }
	n := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.Type().IsDir() { continue }
		n = append(n, e.Name())
	}
	return n, nil
}
func (a sshFSAdapter) ReadFile(path string) ([]byte, error)                 { return a.fs.ReadFile(path) }
func (a sshFSAdapter) WriteFile(path string, data []byte, perm fs.FileMode) error { return a.fs.WriteFile(path, data, perm) }
func (a sshFSAdapter) MkdirAll(path string, perm fs.FileMode) error         { return a.fs.MkdirAll(path, perm) }
func (a sshFSAdapter) Chmod(path string, mode fs.FileMode) error            { return a.fs.Chmod(path, mode) }
func (a sshFSAdapter) UserHomeDir() (string, error)                          { return a.fs.UserHomeDir() }

// nerdfont adapters
type nerdfontFSAdapter struct{ fs runtimepkg.FS }
func (a nerdfontFSAdapter) ReadDir(dir string) ([]string, error) {
	entries, err := a.fs.ReadDir(dir)
	if err != nil { return nil, err }
	n := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.Type().IsDir() { continue }
		n = append(n, e.Name())
	}
	return n, nil
}
func (a nerdfontFSAdapter) ReadFile(path string) ([]byte, error) { return a.fs.ReadFile(path) }

type nerdfontRunnerAdapter struct{ r runtimepkg.Runner }
func (a nerdfontRunnerAdapter) PowerShell(args ...string) (string, error) { return a.r.Output("powershell.exe", args...) }
func (a nerdfontRunnerAdapter) WSLPath(args ...string) (string, error)    { return a.r.Output("wslpath", args...) }

type nerdfontPlatformAdapter struct{ p *platform.Service }
func (a nerdfontPlatformAdapter) IsWSL() bool { return a.p.IsWSL() }

// platform adapters
type platformFSAdapter struct{ fs runtimepkg.FS }
func (a platformFSAdapter) ReadFile(path string) ([]byte, error) { return a.fs.ReadFile(path) }
func (a platformFSAdapter) Stat(path string) (interface{}, error) { return a.fs.Stat(path) }

type platformEnvAdapter struct{ e runtimepkg.Env }
func (a platformEnvAdapter) Getenv(k string) string { return a.e.Getenv(k) }

// user adapters
type userCmdAdapter struct{ r runtimepkg.Runner }
func (a userCmdAdapter) Run(name string, args ...string) error { return a.r.Run(name, args...) }
func (a userCmdAdapter) RunWithStdin(name, stdin string, args ...string) error {
	cmdline := name
	for _, ar := range args { cmdline += " " + ar }
	return a.r.Run("sh", "-c", "printf %s \"$1\" | "+cmdline, "_", stdin)
}

type userFSAdapter struct{ fs runtimepkg.FS }
func (a userFSAdapter) ReadFile(path string) ([]byte, error)                 { return a.fs.ReadFile(path) }
func (a userFSAdapter) WriteFile(path string, data []byte, perm fs.FileMode) error { return a.fs.WriteFile(path, data, perm) }
func (a userFSAdapter) MkdirAll(path string, perm fs.FileMode) error         { return a.fs.MkdirAll(path, perm) }
func (a userFSAdapter) Chmod(path string, mode fs.FileMode) error            { return a.fs.Chmod(path, mode) }

type userLookupAdapter struct{}
func (userLookupAdapter) UserExists(username string) bool {
	_, err := stduser.Lookup(username)
	return err == nil
}
func (userLookupAdapter) HomeDirByUsername(username string) (string, error) {
	u, err := stduser.Lookup(username)
	if err != nil { return "", err }
	return u.HomeDir, nil
}
func (userLookupAdapter) CurrentUsername() string {
	if u, err := stduser.Current(); err == nil && u != nil { return u.Username }
	return ""
}

type userSudoValidator struct{}
func (userSudoValidator) Validate(content string) error { return nil }

// config FS adapter
type configFSAdapter struct{ fs runtimepkg.FS }
func (a configFSAdapter) ReadFile(p string) ([]byte, error)                 { return a.fs.ReadFile(p) }
func (a configFSAdapter) WriteFile(p string, d []byte, perm fs.FileMode) error { return a.fs.WriteFile(p, d, perm) }
func (a configFSAdapter) MkdirAll(p string, perm fs.FileMode) error         { return a.fs.MkdirAll(p, perm) }

func NewProvider() *Provider {
	// Prod runtime deps
	r := runtimepkg.NewRunner(30 * time.Second)
	fs := runtimepkg.NewFS()
	env := runtimepkg.NewEnv()

	// platform DI service
	platSvc := platform.NewService(platformFSAdapter{fs: fs}, platformEnvAdapter{e: env})

	// Services
	fwSvc := firewall.NewService(runnerAdapter{r: r})
	gitSvc := git.NewService(runnerAdapter{r: r})
	sshSvc := ssh.NewService(sshPlatformAdapter{p: platSvc}, sshFSAdapter{fs: fs})
	nerdSvc := nerdfont.NewService(nerdfontPlatformAdapter{p: platSvc}, nerdfontFSAdapter{fs: fs}, nerdfontRunnerAdapter{r: r})
	goSvc := golang.NewService(runnerAdapter{r: r}, struct{ golang.VersionSource }{})
	nodeSvc := nodejs.NewService(nodeRunnerAdapter{r: r}, struct{ nodejs.VersionSource }{})
	pySvc := python.NewService(runnerAdapter{r: r}, struct{ python.VersionSource }{})
	userSvc := user.NewService(userCmdAdapter{r: r}, userFSAdapter{fs: fs}, userLookupAdapter{}, userSudoValidator{})

	// config service
	cfgSvc := config.NewService(configFSAdapter{fs: fs}, fs.UserHomeDir)

	return &Provider{
		User:            userSvc,
		SSH:             sshSvc,
		Git:             gitSvc,
		Firewall:        fwSvc,
		NerdFont:        nerdSvc,
		GoToolchain:     goSvc,
		NodeToolchain:   nodeSvc,
		PythonToolchain: pySvc,
		Platform:        platSvc,
		Config:          cfgSvc,
	}
}
