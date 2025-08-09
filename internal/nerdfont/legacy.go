package nerdfont

import (
	"time"

	"archwsl-tui-configurator/internal/platform"
)

// TODO: remove after DI migration is complete.

type seamFS struct{}

func (seamFS) ReadDir(dir string) ([]string, error) { return readFontDir(dir) }
func (seamFS) ReadFile(path string) ([]byte, error) { return readFileConf(path) }

type seamRunner struct{}

func (seamRunner) PowerShell(args ...string) (string, error) { return runPSCapture(args...) }
func (seamRunner) WSLPath(args ...string) (string, error)    { return runWSLCapture(args...) }

type seamPlatform struct{}

func (seamPlatform) IsWSL() bool { return platform.IsWSL() }

var defaultService = NewService(seamPlatform{}, seamFS{}, seamRunner{})

// Deprecated: prefer constructing nerdfont.Service with DI. This shim will be removed.
func detectNerdFontInstalled() bool { return legacyDetectNerdFontInstalled() }

// keep time import used
var _ = time.Second
