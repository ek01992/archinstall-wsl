package python

import (
    "errors"
    "testing"
)

func TestInstallPythonToolchain_InstallsAllWhenMissing(t *testing.T) {
    origFetch := fetchLatestPythonVersion
    origRun := runCommand
    origCap := runCommandCapture
    t.Cleanup(func() {
        fetchLatestPythonVersion = origFetch
        runCommand = origRun
        runCommandCapture = origCap
    })

    fetchLatestPythonVersion = func() (string, error) { return "3.12.2", nil }

    installedPyenv := false
    installedPipx := false
    pythonSet := false

    runCommand = func(name string, args ...string) error {
        switch name {
        case "pacman":
            if len(args) == 3 && args[0] == "-S" && args[1] == "--noconfirm" && args[2] == "pyenv" {
                installedPyenv = true
                return nil
            }
            if len(args) == 3 && args[0] == "-S" && args[1] == "--noconfirm" && args[2] == "pipx" {
                installedPipx = true
                return nil
            }
            t.Fatalf("unexpected pacman args: %v", args)
        case "pyenv":
            if len(args) == 2 && args[0] == "install" && args[1] == "-s 3.12.2" {
                // combined flag; our implementation could split as separate args; accept either in code
                pythonSet = true
                return nil
            }
            if len(args) == 3 && args[0] == "install" && args[1] == "-s" && args[2] == "3.12.2" {
                return nil
            }
            if len(args) == 2 && args[0] == "global" && args[1] == "3.12.2" {
                pythonSet = true
                return nil
            }
            t.Fatalf("unexpected pyenv args: %v", args)
        default:
            t.Fatalf("unexpected command %q", name)
        }
        return nil
    }

    runCommandCapture = func(name string, args ...string) (string, error) {
        switch name {
        case "pyenv":
            if len(args) == 1 && args[0] == "--version" {
                if installedPyenv {
                    return "pyenv 2.3.24", nil
                }
                return "", errors.New("pyenv not found")
            }
        case "python":
            if len(args) == 1 && args[0] == "--version" {
                if pythonSet {
                    return "Python 3.12.2", nil
                }
                return "", errors.New("python not configured")
            }
        case "pipx":
            if len(args) == 1 && args[0] == "--version" {
                if installedPipx {
                    return "1.5.0", nil
                }
                return "", errors.New("pipx not found")
            }
        }
        t.Fatalf("unexpected capture call: %q %v", name, args)
        return "", nil
    }

    if err := installPythonToolchain(); err != nil {
        t.Fatalf("installPythonToolchain returned error: %v", err)
    }
    if !installedPyenv || !installedPipx || !pythonSet {
        t.Fatalf("expected pyenv, pipx installed and python configured; got pyenv=%v pipx=%v pythonSet=%v", installedPyenv, installedPipx, pythonSet)
    }
}

func TestInstallPythonToolchain_IdempotentWhenUpToDate(t *testing.T) {
    origFetch := fetchLatestPythonVersion
    origRun := runCommand
    origCap := runCommandCapture
    t.Cleanup(func() {
        fetchLatestPythonVersion = origFetch
        runCommand = origRun
        runCommandCapture = origCap
    })

    fetchLatestPythonVersion = func() (string, error) { return "3.12.2", nil }

    runCommand = func(name string, args ...string) error {
        t.Fatalf("no state-changing commands expected; got %q %v", name, args)
        return nil
    }
    runCommandCapture = func(name string, args ...string) (string, error) {
        switch name {
        case "pyenv":
            return "pyenv 2.3.24", nil
        case "python":
            return "Python 3.12.2", nil
        case "pipx":
            return "1.5.0", nil
        }
        return "", nil
    }

    if err := installPythonToolchain(); err != nil {
        t.Fatalf("installPythonToolchain returned error: %v", err)
    }
}

func TestInstallPythonToolchain_UpdatesPythonWhenOutdated(t *testing.T) {
    origFetch := fetchLatestPythonVersion
    origRun := runCommand
    origCap := runCommandCapture
    t.Cleanup(func() {
        fetchLatestPythonVersion = origFetch
        runCommand = origRun
        runCommandCapture = origCap
    })

    fetchLatestPythonVersion = func() (string, error) { return "3.12.2", nil }

    globalCalled := false
    runCommand = func(name string, args ...string) error {
        if name != "pyenv" {
            t.Fatalf("unexpected command %q", name)
        }
        if len(args) == 3 && args[0] == "install" && args[1] == "-s" && args[2] == "3.12.2" {
            return nil
        }
        if len(args) == 2 && args[0] == "global" && args[1] == "3.12.2" {
            globalCalled = true
            return nil
        }
        t.Fatalf("unexpected pyenv args: %v", args)
        return nil
    }

    runCommandCapture = func(name string, args ...string) (string, error) {
        switch name {
        case "pyenv":
            return "pyenv 2.3.24", nil
        case "python":
            if globalCalled {
                return "Python 3.12.2", nil
            }
            return "Python 3.11.1", nil
        case "pipx":
            return "1.5.0", nil
        }
        return "", nil
    }

    if err := installPythonToolchain(); err != nil {
        t.Fatalf("installPythonToolchain returned error: %v", err)
    }
    if !globalCalled {
        t.Fatalf("expected pyenv global to be called to set desired version")
    }
}

func TestInstallPythonToolchain_InstallsPipxWhenMissing(t *testing.T) {
    origFetch := fetchLatestPythonVersion
    origRun := runCommand
    origCap := runCommandCapture
    t.Cleanup(func() {
        fetchLatestPythonVersion = origFetch
        runCommand = origRun
        runCommandCapture = origCap
    })

    fetchLatestPythonVersion = func() (string, error) { return "3.12.2", nil }

    pipxInstalled := false
    runCommand = func(name string, args ...string) error {
        if name == "pacman" && len(args) == 3 && args[0] == "-S" && args[1] == "--noconfirm" && args[2] == "pipx" {
            pipxInstalled = true
            return nil
        }
        t.Fatalf("unexpected runCommand call: %q %v", name, args)
        return nil
    }

    runCommandCapture = func(name string, args ...string) (string, error) {
        switch name {
        case "pyenv":
            return "pyenv 2.3.24", nil
        case "python":
            return "Python 3.12.2", nil
        case "pipx":
            if pipxInstalled {
                return "1.5.0", nil
            }
            return "", errors.New("pipx not found")
        }
        return "", nil
    }

    if err := installPythonToolchain(); err != nil {
        t.Fatalf("installPythonToolchain returned error: %v", err)
    }
    if !pipxInstalled {
        t.Fatalf("expected pipx to be installed when missing")
    }
}

func TestInstallPythonToolchain_FetchFails(t *testing.T) {
    origFetch := fetchLatestPythonVersion
    origRun := runCommand
    t.Cleanup(func() { fetchLatestPythonVersion = origFetch; runCommand = origRun })

    fetchLatestPythonVersion = func() (string, error) { return "", errors.New("fail") }

    runCommand = func(name string, args ...string) error {
        t.Fatalf("no commands should be run when fetch fails")
        return nil
    }

    if err := installPythonToolchain(); err == nil {
        t.Fatalf("expected error when fetch latest python fails")
    }
}
