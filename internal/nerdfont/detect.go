package nerdfont

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"archwsl-tui-configurator/internal/platform"
)

var (
	// enumerateWindowsFontFiles lists font filenames from the Windows Fonts directory via WSL mount.
	enumerateWindowsFontFiles = func() ([]string, error) {
		// Probe common mount roots for Windows C: drive fonts
		if isWSL() {
			candidates := []string{
				"/mnt/c/Windows/Fonts",
				"/c/Windows/Fonts",
				"/mnt/host/c/Windows/Fonts",
			}
			for _, dir := range candidates {
				if names, err := readFontDir(dir); err == nil {
					return names, nil
				}
			}
		}
		// Alternate path probing via wslpath if running under WSL
		if isWSL() {
			if p, err := runWSLCapture("-u", "C:\\Windows\\Fonts"); err == nil {
				path := strings.TrimSpace(p)
				if names, err2 := readFontDir(path); err2 == nil {
					return names, nil
				}
			}
		}
		return nil, os.ErrNotExist
	}

	// readFontDir reads filenames in a directory (seam for testing)
	readFontDir = func(dir string) ([]string, error) {
		entries, err := os.ReadDir(dir)
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

	// runPSCapture runs a PowerShell command (available in WSL environments) and returns stdout.
	runPSCapture = func(args ...string) (string, error) {
		cmd := exec.Command("powershell.exe", args...)
		out, err := cmd.CombinedOutput()
		return string(out), err
	}

	// runWSLCapture runs wslpath or similar WSL utilities
	runWSLCapture = func(args ...string) (string, error) {
		cmd := exec.Command("wslpath", args...)
		out, err := cmd.CombinedOutput()
		return string(out), err
	}

	// isWSL is a seam around platform.IsWSL for tests
	isWSL = func() bool { return platform.IsWSL() }
)

func registryNerdFontPresent() bool {
	if !isWSL() {
		return false
	}
	// Query Windows registry for installed fonts and look for Nerd Font entries
	// HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Fonts
	out, err := runPSCapture("-NoProfile", "-Command", `Get-ItemProperty 'HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Fonts' | Select-Object -ExpandProperty PSObject.Properties | ForEach-Object { $_.Value }`)
	if err != nil || strings.TrimSpace(out) == "" {
		return false
	}
	lower := strings.ToLower(out)
	return strings.Contains(lower, "nerd font") || strings.Contains(lower, "nerdfont")
}

// detectNerdFontInstalled returns true if any installed Windows font filename indicates a Nerd Font.
// The check is case-insensitive and looks for the token "nerd font" in the filename.
func detectNerdFontInstalled() bool {
	files, err := enumerateWindowsFontFiles()
	if err != nil {
		// Best-effort registry fallback when direct enumeration fails
		return registryNerdFontPresent()
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
