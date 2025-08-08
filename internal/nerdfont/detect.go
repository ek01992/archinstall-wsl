package nerdfont

import (
	"bufio"
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
			// 1) Probe wsl.conf automount root, if configured
			if root := strings.TrimSpace(parseWSLAutomountRoot()); root != "" {
				candidate := filepath.Join(root, "c/Windows/Fonts")
				if names, err := readFontDir(candidate); err == nil {
					return names, nil
				}
			}
			// 2) Probe known roots
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

	// readFileConf is a seam for reading configuration files like /etc/wsl.conf
	readFileConf = func(path string) ([]byte, error) { return os.ReadFile(path) }
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

// parseWSLAutomountRoot parses /etc/wsl.conf and returns the automount root, if configured.
func parseWSLAutomountRoot() string {
	data, err := readFileConf("/etc/wsl.conf")
	if err != nil || len(data) == 0 {
		return ""
	}
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	section := ""
	root := ""
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.ToLower(strings.Trim(strings.TrimPrefix(strings.TrimSuffix(line, "]"), "["), " "))
			continue
		}
		if section == "automount" {
			kv := strings.SplitN(line, "=", 2)
			if len(kv) == 2 && strings.TrimSpace(strings.ToLower(kv[0])) == "root" {
				root = strings.TrimSpace(kv[1])
			}
		}
	}
	return strings.TrimSpace(root)
}

// legacyDetectNerdFontInstalled is the previous name to avoid duplicate symbol; legacy.go provides a wrapper.
func legacyDetectNerdFontInstalled() bool {
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
