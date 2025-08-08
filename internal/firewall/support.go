package firewall

// isUfwSupported returns true if ufw appears to be present and responsive.
func isUfwSupported() bool {
	if out, err := runCommandCapture("ufw", "status"); err == nil && out != "" {
		return true
	}
	return false
}
