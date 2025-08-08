package nerdfont

import (
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

func (s *Service) Detect() bool {
	if !s.p.IsWSL() { return false }
	root := parseWSLAutomountRoot()
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
