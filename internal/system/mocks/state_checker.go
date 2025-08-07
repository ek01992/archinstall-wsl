package mocks

import (
	"sync"
)

// MockStateChecker is a configurable mock implementation of the StateChecker interface
type MockStateChecker struct {
	mu sync.RWMutex

	// Configurable responses
	userExistsResponses       map[string]bool
	packageInstalledResponses map[string]bool
	fileContainsResponses     map[string]bool
	isSudoerResponses         map[string]bool
	directoryExistsResponses  map[string]bool
	fileExistsResponses       map[string]bool
	serviceRunningResponses   map[string]bool
	serviceEnabledResponses   map[string]bool

	// Default responses
	defaultUserExists       bool
	defaultPackageInstalled bool
	defaultFileContains     bool
	defaultIsSudoer         bool
	defaultDirectoryExists  bool
	defaultFileExists       bool
	defaultServiceRunning   bool
	defaultServiceEnabled   bool
}

// NewMockStateChecker creates a new MockStateChecker
func NewMockStateChecker() *MockStateChecker {
	return &MockStateChecker{
		userExistsResponses:       make(map[string]bool),
		packageInstalledResponses: make(map[string]bool),
		fileContainsResponses:     make(map[string]bool),
		isSudoerResponses:         make(map[string]bool),
		directoryExistsResponses:  make(map[string]bool),
		fileExistsResponses:       make(map[string]bool),
		serviceRunningResponses:   make(map[string]bool),
		serviceEnabledResponses:   make(map[string]bool),
	}
}

// SetUserExists sets the response for UserExists
func (m *MockStateChecker) SetUserExists(username string, exists bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.userExistsResponses[username] = exists
}

// SetDefaultUserExists sets the default response for UserExists
func (m *MockStateChecker) SetDefaultUserExists(exists bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.defaultUserExists = exists
}

// SetPackageInstalled sets the response for PackageIsInstalled
func (m *MockStateChecker) SetPackageInstalled(pkgName string, installed bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.packageInstalledResponses[pkgName] = installed
}

// SetDefaultPackageInstalled sets the default response for PackageIsInstalled
func (m *MockStateChecker) SetDefaultPackageInstalled(installed bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.defaultPackageInstalled = installed
}

// SetFileContains sets the response for FileContains
func (m *MockStateChecker) SetFileContains(path, content string, contains bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := path + ":" + content
	m.fileContainsResponses[key] = contains
}

// SetDefaultFileContains sets the default response for FileContains
func (m *MockStateChecker) SetDefaultFileContains(contains bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.defaultFileContains = contains
}

// SetIsSudoer sets the response for IsSudoer
func (m *MockStateChecker) SetIsSudoer(username string, isSudoer bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.isSudoerResponses[username] = isSudoer
}

// SetDefaultIsSudoer sets the default response for IsSudoer
func (m *MockStateChecker) SetDefaultIsSudoer(isSudoer bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.defaultIsSudoer = isSudoer
}

// SetDirectoryExists sets the response for DirectoryExists
func (m *MockStateChecker) SetDirectoryExists(path string, exists bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.directoryExistsResponses[path] = exists
}

// SetDefaultDirectoryExists sets the default response for DirectoryExists
func (m *MockStateChecker) SetDefaultDirectoryExists(exists bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.defaultDirectoryExists = exists
}

// SetFileExists sets the response for FileExists
func (m *MockStateChecker) SetFileExists(path string, exists bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.fileExistsResponses[path] = exists
}

// SetDefaultFileExists sets the default response for FileExists
func (m *MockStateChecker) SetDefaultFileExists(exists bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.defaultFileExists = exists
}

// SetServiceRunning sets the response for ServiceIsRunning
func (m *MockStateChecker) SetServiceRunning(serviceName string, running bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.serviceRunningResponses[serviceName] = running
}

// SetDefaultServiceRunning sets the default response for ServiceIsRunning
func (m *MockStateChecker) SetDefaultServiceRunning(running bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.defaultServiceRunning = running
}

// SetServiceEnabled sets the response for ServiceIsEnabled
func (m *MockStateChecker) SetServiceEnabled(serviceName string, enabled bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.serviceEnabledResponses[serviceName] = enabled
}

// SetDefaultServiceEnabled sets the default response for ServiceIsEnabled
func (m *MockStateChecker) SetDefaultServiceEnabled(enabled bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.defaultServiceEnabled = enabled
}

// UserExists implements the StateChecker interface
func (m *MockStateChecker) UserExists(username string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if exists, found := m.userExistsResponses[username]; found {
		return exists, nil
	}

	return m.defaultUserExists, nil
}

// PackageIsInstalled implements the StateChecker interface
func (m *MockStateChecker) PackageIsInstalled(pkgName string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if installed, found := m.packageInstalledResponses[pkgName]; found {
		return installed, nil
	}

	return m.defaultPackageInstalled, nil
}

// FileContains implements the StateChecker interface
func (m *MockStateChecker) FileContains(path, content string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := path + ":" + content
	if contains, found := m.fileContainsResponses[key]; found {
		return contains, nil
	}

	return m.defaultFileContains, nil
}

// IsSudoer implements the StateChecker interface
func (m *MockStateChecker) IsSudoer(username string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if isSudoer, found := m.isSudoerResponses[username]; found {
		return isSudoer, nil
	}

	return m.defaultIsSudoer, nil
}

// DirectoryExists implements the StateChecker interface
func (m *MockStateChecker) DirectoryExists(path string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if exists, found := m.directoryExistsResponses[path]; found {
		return exists, nil
	}

	return m.defaultDirectoryExists, nil
}

// FileExists implements the StateChecker interface
func (m *MockStateChecker) FileExists(path string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if exists, found := m.fileExistsResponses[path]; found {
		return exists, nil
	}

	return m.defaultFileExists, nil
}

// ServiceIsRunning implements the StateChecker interface
func (m *MockStateChecker) ServiceIsRunning(serviceName string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if running, found := m.serviceRunningResponses[serviceName]; found {
		return running, nil
	}

	return m.defaultServiceRunning, nil
}

// ServiceIsEnabled implements the StateChecker interface
func (m *MockStateChecker) ServiceIsEnabled(serviceName string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if enabled, found := m.serviceEnabledResponses[serviceName]; found {
		return enabled, nil
	}

	return m.defaultServiceEnabled, nil
}
