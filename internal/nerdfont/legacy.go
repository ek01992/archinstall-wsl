package nerdfont

// Deprecated: prefer constructing nerdfont.Service with DI. This shim will be removed.
func detectNerdFontInstalled() bool { return legacyDetectNerdFontInstalled() }
