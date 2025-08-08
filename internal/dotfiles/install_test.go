package dotfiles

import (
    "io/fs"
    "path/filepath"
    "strings"
    "testing"
)

func TestInstallDotfiles_CloneAndSymlink(t *testing.T) {
    origHome := getUserHomeDir
    origExists := pathExists
    origRun := runCommand
    origList := listFiles
    origSymlink := symlink
    origLstat := lstat
    origReadlink := readlink

    t.Cleanup(func() {
        getUserHomeDir = origHome
        pathExists = origExists
        runCommand = origRun
        listFiles = origList
        symlink = origSymlink
        lstat = origLstat
        readlink = origReadlink
    })

    getUserHomeDir = func() (string, error) { return "/home/alice", nil }

    // Repo absent initially
    pathExists = func(path string) bool {
        return false
    }

    gitCloned := false
    runCommand = func(name string, args ...string) error {
        if name != "git" {
            t.Fatalf("expected git, got %q", name)
        }
        if len(args) != 5 || args[0] != "clone" || args[1] != "--depth" || args[2] != "1" || args[3] != "https://example.com/dotfiles.git" || !strings.HasSuffix(args[4], "/.dotfiles") {
            t.Fatalf("unexpected git clone args: %v", args)
        }
        gitCloned = true
        return nil
    }

    listFiles = func(dir string) ([]string, error) {
        if !strings.HasSuffix(dir, "/.dotfiles") {
            t.Fatalf("unexpected list dir %q", dir)
        }
        return []string{"zshrc", ".gitconfig", "README.md"}, nil
    }

    lstat = func(path string) (fs.FileMode, error) {
        return 0, fs.ErrNotExist
    }
    readlink = func(path string) (string, error) { return "", fs.ErrNotExist }

    links := map[string]string{}
    symlink = func(oldname, newname string) error {
        links[newname] = oldname
        return nil
    }

    if err := installDotfiles("https://example.com/dotfiles.git"); err != nil {
        t.Fatalf("installDotfiles returned error: %v", err)
    }

    if !gitCloned {
        t.Fatalf("expected git clone to be called")
    }

    // We expect zshrc -> ~/.zshrc and .gitconfig -> ~/.gitconfig
    if links["/home/alice/.zshrc"] != "/home/alice/.dotfiles/zshrc" {
        t.Fatalf("unexpected symlink for zshrc: %q", links["/home/alice/.zshrc"])
    }
    if links["/home/alice/.gitconfig"] != "/home/alice/.dotfiles/.gitconfig" {
        t.Fatalf("unexpected symlink for gitconfig: %q", links["/home/alice/.gitconfig"])
    }
    if _, ok := links[filepath.Join("/home/alice", "README.md")]; ok {
        t.Fatalf("should not symlink README.md")
    }
}

func TestInstallDotfiles_Idempotent_SkipWhenSymlinksCorrect(t *testing.T) {
    origHome := getUserHomeDir
    origExists := pathExists
    origRun := runCommand
    origList := listFiles
    origSymlink := symlink
    origLstat := lstat
    origReadlink := readlink

    t.Cleanup(func() {
        getUserHomeDir = origHome
        pathExists = origExists
        runCommand = origRun
        listFiles = origList
        symlink = origSymlink
        lstat = origLstat
        readlink = origReadlink
    })

    getUserHomeDir = func() (string, error) { return "/home/bob", nil }
    pathExists = func(path string) bool { return true }

    runCommand = func(name string, args ...string) error {
        t.Fatalf("did not expect git clone when repo exists")
        return nil
    }

    listFiles = func(dir string) ([]string, error) { return []string{"zshrc"}, nil }

    // Target already correct symlink
    lstat = func(path string) (fs.FileMode, error) { return fs.ModeSymlink, nil }
    readlink = func(path string) (string, error) { return "/home/bob/.dotfiles/zshrc", nil }

    calledSymlink := false
    symlink = func(oldname, newname string) error { calledSymlink = true; return nil }

    if err := installDotfiles("https://example.com/dotfiles.git"); err != nil {
        t.Fatalf("installDotfiles returned error: %v", err)
    }
    if calledSymlink {
        t.Fatalf("did not expect symlink to be recreated when already correct")
    }
}

func TestInstallDotfiles_DefaultsWhenNoRepoURL(t *testing.T) {
    origHome := getUserHomeDir
    origExists := pathExists
    origRun := runCommand
    origWrite := writeFile

    t.Cleanup(func() {
        getUserHomeDir = origHome
        pathExists = origExists
        runCommand = origRun
        writeFile = origWrite
    })

    getUserHomeDir = func() (string, error) { return "/home/carol", nil }

    runCommand = func(name string, args ...string) error {
        t.Fatalf("no commands should be run when repoURL empty")
        return nil
    }
    pathExists = func(path string) bool { return false }

    var wrotePath, wroteContent string
    writeFile = func(path string, data []byte, perm fs.FileMode) error {
        wrotePath = path
        wroteContent = string(data)
        return nil
    }

    if err := installDotfiles("   \t\n"); err != nil {
        t.Fatalf("installDotfiles returned error: %v", err)
    }
    if wrotePath != "/home/carol/.zshrc" {
        t.Fatalf("expected default .zshrc to be written at /home/carol/.zshrc, got %q", wrotePath)
    }
    if !strings.Contains(wroteContent, "export ZSH=\"$HOME/.oh-my-zsh\"") {
        t.Fatalf("expected default zshrc content to include oh-my-zsh setup; got %q", wroteContent)
    }
}
