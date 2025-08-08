package version

import (
	"regexp"
	"testing"
)

func TestVersionIsSemver(t *testing.T) {
	if Version == "" {
		t.Fatalf("Version must not be empty")
	}
	semverPattern := `^v?\d+\.\d+\.\d+(-[A-Za-z0-9-.]+)?$`
	re := regexp.MustCompile(semverPattern)
	if !re.MatchString(Version) {
		t.Fatalf("Version %q does not match semver pattern %q", Version, semverPattern)
	}
}
