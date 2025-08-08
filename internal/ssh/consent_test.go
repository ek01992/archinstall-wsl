package ssh

import "testing"

func TestImportSSHKeysWithConsent_Decline_NoCalls(t *testing.T) {
	// Ensure underlying import isn't called
	origImport := importSSHKeys
	origCan := canEditHostFiles
	defer func() { importSSHKeys = origImport; canEditHostFiles = origCan }()
	importSSHKeys = func(hostPath string) error {
		t.Fatalf("import should not be called when consent is false")
		return nil
	}
	canEditHostFiles = func() bool { t.Fatalf("canEditHostFiles should not be called when consent is false"); return false }
	if err := importSSHKeysWithConsent("/mnt/c/Users/Alice/.ssh", false); err == nil {
		t.Fatalf("expected error when consent is false")
	}
}

func TestImportSSHKeysWithConsent_NoMount_Error(t *testing.T) {
	origCan := canEditHostFiles
	defer func() { canEditHostFiles = origCan }()
	canEditHostFiles = func() bool { return false }
	if err := importSSHKeysWithConsent("/mnt/c/Users/Alice/.ssh", true); err == nil {
		t.Fatalf("expected error when host files are not accessible")
	}
}

func TestImportSSHKeysWithConsent_Succeeds(t *testing.T) {
	origImport := importSSHKeys
	origCan := canEditHostFiles
	defer func() { importSSHKeys = origImport; canEditHostFiles = origCan }()
	called := false
	importSSHKeys = func(hostPath string) error {
		called = true
		if hostPath != "/mnt/c/Users/Alice/.ssh" { t.Fatalf("unexpected hostPath: %q", hostPath) }
		return nil
	}
	canEditHostFiles = func() bool { return true }
	if err := importSSHKeysWithConsent("/mnt/c/Users/Alice/.ssh", true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called { t.Fatalf("expected underlying import to be called") }
}
