package pacman

import "strings"

// isPackageInstalledWithQuery is a testable variant that accepts the query function.
func isPackageInstalledWithQuery(pkg string, query func(string) error) bool {
	pkg = strings.TrimSpace(pkg)
	if pkg == "" {
		return false
	}
	return query(pkg) == nil
}
