package app

import "testing"

func TestNewProvider_ConstructsAllServices(t *testing.T) {
	p := NewProvider()
	if p == nil { t.Fatalf("provider is nil") }
	if p.Platform == nil { t.Fatalf("platform not constructed") }
	if p.User == nil { t.Fatalf("user not constructed") }
	if p.SSH == nil { t.Fatalf("ssh not constructed") }
	if p.Git == nil { t.Fatalf("git not constructed") }
	if p.Firewall == nil { t.Fatalf("firewall not constructed") }
	if p.NerdFont == nil { t.Fatalf("nerdfont not constructed") }
	if p.GoToolchain == nil { t.Fatalf("go toolchain not constructed") }
	if p.NodeToolchain == nil { t.Fatalf("node toolchain not constructed") }
	if p.PythonToolchain == nil { t.Fatalf("python toolchain not constructed") }
}
