package pacman

import (
	"context"
	"io"
	"os/exec"
	"strings"
	"time"
)

// queryLocalPackage is a test seam to allow mocking the pacman invocation.
var queryLocalPackage = func(pkg string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "pacman", "-Q", pkg)
	// We only care about exit status; suppress output to avoid noise in WSL.
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	return cmd.Run()
}

// isPackageInstalled returns true if the given package appears installed locally
// according to `pacman -Q <pkg>`. Empty or whitespace-only names return false.
func isPackageInstalled(pkg string) bool {
	pkg = strings.TrimSpace(pkg)
	if pkg == "" {
		return false
	}
	return queryLocalPackage(pkg) == nil
}
