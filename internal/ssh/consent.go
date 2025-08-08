package ssh

import (
	"errors"

	"archwsl-tui-configurator/internal/platform"
)

var canEditHostFiles = func() bool { return platform.CanEditHostFiles() }
var importSSHKeys = func(hostPath string) error { return importSSHKeysFromWindows(hostPath) }

// importSSHKeysWithConsent imports keys only if the caller explicitly consents and host files are editable.
func importSSHKeysWithConsent(hostPath string, consent bool) error {
	if !consent {
		return errors.New("ssh key import requires explicit consent")
	}
	if !canEditHostFiles() {
		return errors.New("host files not accessible (WSL mount missing)")
	}
	return importSSHKeys(hostPath)
}
