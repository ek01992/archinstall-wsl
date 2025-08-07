package system

// StateChecker defines the interface for checking system state
type StateChecker interface {
	// UserExists checks if a user exists in the system
	UserExists(username string) (bool, error)

	// PackageIsInstalled checks if a package is installed
	PackageIsInstalled(pkgName string) (bool, error)

	// FileContains checks if a file contains specific content
	FileContains(path, content string) (bool, error)

	// IsSudoer checks if a user has sudo privileges
	IsSudoer(username string) (bool, error)

	// DirectoryExists checks if a directory exists
	DirectoryExists(path string) (bool, error)

	// FileExists checks if a file exists
	FileExists(path string) (bool, error)

	// ServiceIsRunning checks if a systemd service is running
	ServiceIsRunning(serviceName string) (bool, error)

	// ServiceIsEnabled checks if a systemd service is enabled
	ServiceIsEnabled(serviceName string) (bool, error)
}
