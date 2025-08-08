package nerdfont

import (
    "os"
    "path/filepath"
    "strings"
)

var (
    // enumerateWindowsFontFiles lists font filenames from the Windows Fonts directory via WSL mount.
    enumerateWindowsFontFiles = func() ([]string, error) {
        fontsDir := "/mnt/c/Windows/Fonts"
        entries, err := os.ReadDir(fontsDir)
        if err != nil {
            return nil, err
        }
        names := make([]string, 0, len(entries))
        for _, e := range entries {
            if e.IsDir() {
                continue
            }
            names = append(names, e.Name())
        }
        return names, nil
    }
)

// detectNerdFontInstalled returns true if any installed Windows font filename indicates a Nerd Font.
// The check is case-insensitive and looks for the token "nerd font" in the filename.
func detectNerdFontInstalled() bool {
    files, err := enumerateWindowsFontFiles()
    if err != nil {
        return false
    }
    for _, name := range files {
        lower := strings.ToLower(filepath.Base(name))
        if strings.Contains(lower, "nerd font") {
            return true
        }
        // Also consider newer naming like "NerdFont-" (no space)
        if strings.Contains(lower, "nerdfont") {
            return true
        }
        // Some distributions use "NF" suffix; be conservative and require explicit phrase to avoid false positives.
    }
    return false
}
