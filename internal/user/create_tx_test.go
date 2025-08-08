package user

import (
	"errors"
	"io/fs"
	"strings"
	"testing"
)

func TestCreateUserTx_RollsBackOnFailure(t *testing.T) {
	origRun := runCommand
	origRunStdin := runCommandWithStdin
	origRead := readFile
	origExist := lookupUserByName

	t.Cleanup(func() { runCommand = origRun; runCommandWithStdin = origRunStdin; readFile = origRead; lookupUserByName = origExist })

	// User does not exist initially
	lookupUserByName = func(name string) (any, error) { return nil, errors.New("no") }

	userdelCalled := false
	rmCalled := false
	runCommand = func(name string, args ...string) error {
		if name == "useradd" {
			return nil
		}
		if name == "usermod" || name == "gpasswd" {
			return nil
		}
		if name == "userdel" {
			userdelCalled = true
			return nil
		}
		if name == "rm" {
			rmCalled = true
			return nil
		}
		return nil
	}

	// Force failure at password setting to ensure useradd ran
	runCommandWithStdin = func(name string, stdin string, args ...string) error {
		if name == "chpasswd" {
			return errors.New("boom")
		}
		return nil
	}

	readFile = func(path string) ([]byte, error) {
		if strings.Contains(path, "/etc/sudoers.d/") {
			return nil, fs.ErrNotExist
		}
		return nil, fs.ErrNotExist
	}

	if err := createUserTx("alice", "pw"); err == nil {
		t.Fatalf("expected error from createUserTx when create fails")
	}
	if !userdelCalled || !rmCalled {
		t.Fatalf("expected rollback to attempt userdel and rm; got userdel=%v rm=%v", userdelCalled, rmCalled)
	}
}

func TestCreateUserTx_SuccessDoesNotRollback(t *testing.T) {
	origRun := runCommand
	origRunStdin := runCommandWithStdin
	origRead := readFile
	origWrite := writeFile
	origExist := lookupUserByName

	t.Cleanup(func() { runCommand = origRun; runCommandWithStdin = origRunStdin; readFile = origRead; writeFile = origWrite; lookupUserByName = origExist })

	lookupUserByName = func(name string) (any, error) { return nil, errors.New("no") }

	userdelCalled := false
	runCommand = func(name string, args ...string) error {
		if name == "userdel" {
			userdelCalled = true
		}
		return nil
	}
	runCommandWithStdin = func(name string, stdin string, args ...string) error { return nil }
	readFile = func(path string) ([]byte, error) {
		if strings.Contains(path, "/etc/sudoers.d/") {
			return []byte("%wheel ALL=(ALL) NOPASSWD: ALL\n"), nil
		}
		return nil, fs.ErrNotExist
	}
	writeFile = func(path string, data []byte, perm fs.FileMode) error { return nil }

	if err := createUserTx("bob", "pw"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if userdelCalled {
		t.Fatalf("should not rollback on success")
	}
}
