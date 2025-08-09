package pacman

import (
	"context"
	"io"
	"os/exec"
	"strings"
	"time"
)

// queryLocalPackage executes `pacman -Q <pkg>` and returns the command error.
func queryLocalPackage(pkg string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "pacman", "-Q", pkg)
	// We only care about exit status; suppress output to avoid noise in WSL.
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	return cmd.Run()
}

// isPackageInstalledWithQuery is a testable variant that accepts the query function.
func isPackageInstalledWithQuery(pkg string, query func(string) error) bool {
	pkg = strings.TrimSpace(pkg)
	if pkg == "" {
		return false
	}
	return query(pkg) == nil
}

// isPackageInstalled returns true if the given package appears installed locally
// according to `pacman -Q <pkg>`. Empty or whitespace-only names return false.
func isPackageInstalled(pkg string) bool {
	return isPackageInstalledWithQuery(pkg, queryLocalPackage)
}
