package nerdfont

import (
	"bufio"
	"path/filepath"
	"strings"
)

type Platform interface { IsWSL() bool }

type FS interface {
	ReadDir(dir string) ([]string, error)
	ReadFile(path string) ([]byte, error)
}

type Runner interface {
	PowerShell(args ...string) (string, error)
	WSLPath(args ...string) (string, error)
}

type Service struct { p Platform; fs FS; r Runner }

func NewService(p Platform, fs FS, r Runner) *Service { return &Service{p: p, fs: fs, r: r} }

func parseWSLAutomountRootFromFS(fs FS) string {
	data, err := fs.ReadFile("/etc/wsl.conf")
	if err != nil || len(data) == 0 { return "" }
	s := bufio.NewScanner(strings.NewReader(string(data)))
	section := ""
	root := ""
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") { continue }
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

func (s *Service) Detect() bool {
	if !s.p.IsWSL() { return false }
	root := parseWSLAutomountRootFromFS(s.fs)
	candidates := []string{}
	if strings.TrimSpace(root) != "" { candidates = append(candidates, filepath.Join(root, "c/Windows/Fonts")) }
	candidates = append(candidates, "/mnt/c/Windows/Fonts", "/c/Windows/Fonts", "/mnt/host/c/Windows/Fonts")
	for _, dir := range candidates {
		if names, err := s.fs.ReadDir(dir); err == nil {
			for _, name := range names {
				lower := strings.ToLower(filepath.Base(name))
				if strings.Contains(lower, "nerd font") || strings.Contains(lower, "nerdfont") { return true }
			}
		}
	}
	if p, err := s.r.WSLPath("-u", "C:\\Windows\\Fonts"); err == nil {
		if names, err2 := s.fs.ReadDir(strings.TrimSpace(p)); err2 == nil {
			for _, name := range names {
				lower := strings.ToLower(filepath.Base(name))
				if strings.Contains(lower, "nerd font") || strings.Contains(lower, "nerdfont") { return true }
			}
		}
	}
	out, err := s.r.PowerShell("-NoProfile", "-Command", `Get-ItemProperty 'HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Fonts' | Select-Object -ExpandProperty PSObject.Properties | ForEach-Object { $_.Value }`)
	if err != nil || strings.TrimSpace(out) == "" { return false }
	lower := strings.ToLower(out)
	return strings.Contains(lower, "nerd font") || strings.Contains(lower, "nerdfont")
}
