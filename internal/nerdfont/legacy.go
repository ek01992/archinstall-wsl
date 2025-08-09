package nerdfont

import (
	"time"
)

// Deprecated: prefer constructing nerdfont.Service with DI. This shim will be removed.
func detectNerdFontInstalled() bool { return legacyDetectNerdFontInstalled() }

// keep time import used
var _ = time.Second
