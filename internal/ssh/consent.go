package ssh

import (
	"fmt"

	"archwsl-tui-configurator/internal/platform"
)

var canEditHostFiles = func() bool { return platform.CanEditHostFiles() }
var importSSHKeys = func(hostPath string) error { return importSSHKeysFromWindows(hostPath) }

// importSSHKeysWithConsent imports keys only if the caller explicitly consents and host files are editable.
func importSSHKeysWithConsent(hostPath string, consent bool) error {
	if !consent {
		return fmt.Errorf("ssh key import: explicit consent required")
	}
	if !canEditHostFiles() {
		return fmt.Errorf("ssh key import: host files not accessible (WSL mount missing)")
	}
	return importSSHKeys(hostPath)
}
