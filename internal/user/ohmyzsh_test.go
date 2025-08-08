package user

import (
	"io/fs"
	"strings"
	"testing"
)

func TestInstallOhMyZsh_ClonesIfMissing_WritesZshrcAndVerifies(t *testing.T) {
	origRun := runCommand
	origRead := readFile
	origWrite := writeFile
	origHomeByUser := getHomeDirByUsername
	origPathExists := pathExists

	t.Cleanup(func() {
		runCommand = origRun
		readFile = origRead
		writeFile = origWrite
		getHomeDirByUsername = origHomeByUser
		pathExists = origPathExists
	})

	username := "alice"
	theme := "agnoster"
	plugins := []string{"git", "fzf", "z"}

	getHomeDirByUsername = func(u string) (string, error) {
		if u != username {
			t.Fatalf("expected username %q, got %q", username, u)
		}
		return "/home/alice", nil
	}

	// oh-my-zsh missing initially
	pathExists = func(path string) bool {
		if strings.HasSuffix(path, "/.oh-my-zsh") {
			return false
		}
		return false
	}

	var cloned bool
	runCommand = func(name string, args ...string) error {
		if name == "git" && len(args) >= 5 && args[0] == "clone" && args[1] == "--depth" && args[2] == "1" && strings.Contains(args[3], "ohmyzsh/ohmyzsh.git") && strings.HasSuffix(args[4], "/.oh-my-zsh") {
			cloned = true
			return nil
		}
		t.Fatalf("unexpected command %q %v", name, args)
		return nil
	}

	// First read: .zshrc does not exist; After write, subsequent read returns written content
	var zshrcPath string
	var written string
	readFile = func(path string) ([]byte, error) {
		if strings.HasSuffix(path, "/.zshrc") {
			zshrcPath = path
			if written == "" {
				return nil, fs.ErrNotExist
			}
			return []byte(written), nil
		}
		t.Fatalf("unexpected readFile path: %q", path)
		return nil, nil
	}

	writeFile = func(path string, data []byte, perm fs.FileMode) error {
		if path != "/home/alice/.zshrc" {
			t.Fatalf("unexpected write path: %q", path)
		}
		if perm != 0o644 {
			t.Fatalf("expected 0644 perm for .zshrc")
		}
		written = string(data)
		return nil
	}

	if err := installOhMyZsh(username, theme, plugins); err != nil {
		t.Fatalf("installOhMyZsh returned error: %v", err)
	}

	if !cloned {
		t.Fatalf("expected oh-my-zsh to be cloned when missing")
	}
	if zshrcPath != "/home/alice/.zshrc" {
		t.Fatalf("expected .zshrc at /home/alice/.zshrc; got %q", zshrcPath)
	}
	if !strings.Contains(written, "export ZSH=\"$HOME/.oh-my-zsh\"") {
		t.Fatalf(".zshrc missing ZSH export: %q", written)
	}
	if !strings.Contains(written, "ZSH_THEME=\"agnoster\"") {
		t.Fatalf(".zshrc missing theme: %q", written)
	}
	if !strings.Contains(written, "plugins=(git fzf z)") {
		t.Fatalf(".zshrc missing plugins list: %q", written)
	}
	if !strings.Contains(written, "source $ZSH/oh-my-zsh.sh") {
		t.Fatalf(".zshrc missing source line: %q", written)
	}
}

func TestInstallOhMyZsh_Idempotent_NoRewriteWhenIdentical(t *testing.T) {
	origRun := runCommand
	origRead := readFile
	origWrite := writeFile
	origHomeByUser := getHomeDirByUsername
	origPathExists := pathExists

	t.Cleanup(func() {
		runCommand = origRun
		readFile = origRead
		writeFile = origWrite
		getHomeDirByUsername = origHomeByUser
		pathExists = origPathExists
	})

	username := "bob"
	getHomeDirByUsername = func(u string) (string, error) { return "/home/bob", nil }
	pathExists = func(path string) bool { return true }

	existing := "export ZSH=\"$HOME/.oh-my-zsh\"\nZSH_THEME=\"robbyrussell\"\nplugins=(git z)\nsource $ZSH/oh-my-zsh.sh\n"
	readFile = func(path string) ([]byte, error) { return []byte(existing), nil }

	wrote := false
	writeFile = func(path string, data []byte, perm fs.FileMode) error { wrote = true; return nil }

	if err := installOhMyZsh(username, "robbyrussell", []string{"git", "z"}); err != nil {
		t.Fatalf("installOhMyZsh returned error: %v", err)
	}
	if wrote {
		t.Fatalf("did not expect rewrite when .zshrc already matches desired state")
	}
}

func TestInstallOhMyZsh_UpdatesWhenDifferent(t *testing.T) {
	origRun := runCommand
	origRead := readFile
	origWrite := writeFile
	origHomeByUser := getHomeDirByUsername
	origPathExists := pathExists

	t.Cleanup(func() {
		runCommand = origRun
		readFile = origRead
		writeFile = origWrite
		getHomeDirByUsername = origHomeByUser
		pathExists = origPathExists
	})

	username := "carol"
	getHomeDirByUsername = func(u string) (string, error) { return "/home/carol", nil }
	pathExists = func(path string) bool { return true }

	existing := "export ZSH=\"$HOME/.oh-my-zsh\"\nZSH_THEME=\"old\"\nplugins=(git)\nsource $ZSH/oh-my-zsh.sh\n"
	var written string
	readFile = func(path string) ([]byte, error) {
		if strings.HasSuffix(path, "/.zshrc") {
			if written == "" {
				return []byte(existing), nil
			}
			return []byte(written), nil
		}
		return nil, fs.ErrNotExist
	}
	writeFile = func(path string, data []byte, perm fs.FileMode) error { written = string(data); return nil }

	if err := installOhMyZsh(username, "newtheme", []string{"git", "fzf"}); err != nil {
		t.Fatalf("installOhMyZsh returned error: %v", err)
	}
	if !strings.Contains(written, "ZSH_THEME=\"newtheme\"") || !strings.Contains(written, "plugins=(git fzf)") {
		t.Fatalf("expected updated theme/plugins, got %q", written)
	}
}

func TestInstallOhMyZsh_EmptyUsernameError_NoCalls(t *testing.T) {
	origRun := runCommand
	origRead := readFile
	origWrite := writeFile
	origHomeByUser := getHomeDirByUsername
	origPathExists := pathExists

	t.Cleanup(func() {
		runCommand = origRun
		readFile = origRead
		writeFile = origWrite
		getHomeDirByUsername = origHomeByUser
		pathExists = origPathExists
	})

	runCommand = func(name string, args ...string) error { t.Fatalf("no commands expected"); return nil }
	readFile = func(path string) ([]byte, error) { t.Fatalf("no reads expected"); return nil, nil }
	writeFile = func(path string, data []byte, perm fs.FileMode) error { t.Fatalf("no writes expected"); return nil }
	getHomeDirByUsername = func(u string) (string, error) { t.Fatalf("no home lookup expected"); return "", nil }
	pathExists = func(path string) bool { t.Fatalf("no path checks expected"); return false }

	if err := installOhMyZsh(" \t\n ", "theme", []string{"git"}); err == nil {
		t.Fatalf("expected error for empty username")
	}
}
