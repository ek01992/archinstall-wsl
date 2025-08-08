package nodejs

import (
    "errors"
    "testing"
)

func TestInstallNodeToolchain_InstallsWhenMissing(t *testing.T) {
    origFetch := fetchLatestNodeLTS
    origRun := runCommand
    origCap := runCommandCapture
    t.Cleanup(func() { fetchLatestNodeLTS = origFetch; runCommand = origRun; runCommandCapture = origCap })

    fetchLatestNodeLTS = func() (string, error) { return "v20.16.0", nil }

    nvmInstalled := false
    nodeInstalled := false

    runCommand = func(name string, args ...string) error {
        switch name {
        case "pacman":
            if len(args) == 3 && args[0] == "-S" && args[1] == "--noconfirm" && args[2] == "nvm" {
                nvmInstalled = true
                return nil
            }
            t.Fatalf("unexpected pacman args: %v", args)
        case "nvm":
            if len(args) == 2 && args[0] == "install" && args[1] == "v20.16.0" {
                return nil
            }
            if len(args) == 3 && args[0] == "alias" && args[1] == "default" && args[2] == "v20.16.0" {
                nodeInstalled = true
                return nil
            }
            t.Fatalf("unexpected nvm args: %v", args)
        default:
            t.Fatalf("unexpected command %q", name)
        }
        return nil
    }

    runCommandCapture = func(name string, args ...string) (string, error) {
        switch name {
        case "nvm":
            if len(args) == 1 && args[0] == "--version" {
                if nvmInstalled {
                    return "0.39.7", nil
                }
                return "", errors.New("nvm not found")
            }
        case "node":
            if len(args) == 1 && args[0] == "-v" {
                if nodeInstalled {
                    return "v20.16.0", nil
                }
                return "", errors.New("node not found")
            }
        }
        t.Fatalf("unexpected capture: %q %v", name, args)
        return "", nil
    }

    if err := installNodeToolchain(); err != nil {
        t.Fatalf("installNodeToolchain returned error: %v", err)
    }
}

func TestInstallNodeToolchain_IdempotentWhenUpToDate(t *testing.T) {
    origFetch := fetchLatestNodeLTS
    origRun := runCommand
    origCap := runCommandCapture
    t.Cleanup(func() { fetchLatestNodeLTS = origFetch; runCommand = origRun; runCommandCapture = origCap })

    fetchLatestNodeLTS = func() (string, error) { return "v20.16.0", nil }

    runCommand = func(name string, args ...string) error {
        t.Fatalf("no state-changing commands expected; got %q %v", name, args)
        return nil
    }
    runCommandCapture = func(name string, args ...string) (string, error) {
        switch name {
        case "nvm":
            return "0.39.7", nil
        case "node":
            return "v20.16.0", nil
        }
        return "", nil
    }

    if err := installNodeToolchain(); err != nil {
        t.Fatalf("installNodeToolchain returned error: %v", err)
    }
}

func TestInstallNodeToolchain_UpdatesWhenOutdated(t *testing.T) {
    origFetch := fetchLatestNodeLTS
    origRun := runCommand
    origCap := runCommandCapture
    t.Cleanup(func() { fetchLatestNodeLTS = origFetch; runCommand = origRun; runCommandCapture = origCap })

    fetchLatestNodeLTS = func() (string, error) { return "v20.16.0", nil }

    aliasCalled := false

    runCommand = func(name string, args ...string) error {
        if name != "nvm" {
            t.Fatalf("unexpected command %q", name)
        }
        if len(args) == 2 && args[0] == "install" && args[1] == "v20.16.0" {
            return nil
        }
        if len(args) == 3 && args[0] == "alias" && args[1] == "default" && args[2] == "v20.16.0" {
            aliasCalled = true
            return nil
        }
        t.Fatalf("unexpected nvm args: %v", args)
        return nil
    }

    returnedNew := false
    runCommandCapture = func(name string, args ...string) (string, error) {
        switch name {
        case "nvm":
            return "0.39.7", nil
        case "node":
            if returnedNew {
                return "v20.16.0", nil
            }
            returnedNew = true
            return "v18.19.0", nil
        }
        return "", nil
    }

    if err := installNodeToolchain(); err != nil {
        t.Fatalf("installNodeToolchain returned error: %v", err)
    }
    if !aliasCalled {
        t.Fatalf("expected nvm alias default to be called")
    }
}

func TestInstallNodeToolchain_FetchFails(t *testing.T) {
    origFetch := fetchLatestNodeLTS
    origRun := runCommand
    t.Cleanup(func() { fetchLatestNodeLTS = origFetch; runCommand = origRun })

    fetchLatestNodeLTS = func() (string, error) { return "", errors.New("fail") }

    runCommand = func(name string, args ...string) error {
        t.Fatalf("no commands should be run when fetch fails")
        return nil
    }

    if err := installNodeToolchain(); err == nil {
        t.Fatalf("expected error when fetch latest LTS fails")
    }
}
