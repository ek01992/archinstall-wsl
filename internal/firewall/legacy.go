package firewall

import (
	"time"

	runtimepkg "archwsl-tui-configurator/internal/runtime"
)

// TODO: remove this legacy shim after DI migration completes.

// defaultService is a package-level Service built from production deps for callers that
// have not migrated to DI yet.
var defaultService = NewService(runtimeRunnerAdapter{r: runtimepkg.NewRunner(10 * time.Second)})

type runtimeRunnerAdapter struct{ r runtimepkg.Runner }

func (a runtimeRunnerAdapter) Run(name string, args ...string) error            { return a.r.Run(name, args...) }
func (a runtimeRunnerAdapter) Output(name string, args ...string) (string, error) { return a.r.Output(name, args...) }

// Deprecated: prefer constructing firewall.Service with DI. This shim will be removed.
func Configure() error { return defaultService.Configure() }

// Deprecated: prefer constructing firewall.Service with DI. This shim will be removed.
func ConfigureTx() error { return defaultService.ConfigureTx() }
