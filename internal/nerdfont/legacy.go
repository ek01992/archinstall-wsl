package nerdfont

import (
	"time"
)

// TODO: remove after DI migration is complete.

// Deprecated: prefer constructing nerdfont.Service with DI. This shim will be removed.
func detectNerdFontInstalled() bool { return legacyDetectNerdFontInstalled() }

// keep time import used
var _ = time.Second
