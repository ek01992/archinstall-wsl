package git

import (
	"strings"
	"sync"
	"testing"
)

type gRunner struct{ mu sync.Mutex; kv map[string]string }
func (g *gRunner) Run(name string, args ...string) error {
	g.mu.Lock(); defer g.mu.Unlock()
	if g.kv==nil { g.kv = map[string]string{} }
	if len(args) >= 4 && args[0]=="config" && args[1]=="--global" && args[2]=="user.name" { g.kv["user.name"] = args[3] }
	if len(args) >= 4 && args[0]=="config" && args[1]=="--global" && args[2]=="user.email" { g.kv["user.email"] = args[3] }
	if len(args) >= 4 && args[0]=="config" && args[1]=="--global" && args[2]=="--unset" { delete(g.kv, args[3]) }
	return nil
}
func (g *gRunner) Output(name string, args ...string) (string, error) {
	g.mu.Lock(); defer g.mu.Unlock()
	if len(args) >= 5 && args[0]=="config" && args[1]=="--global" && args[2]=="--get" {
		return g.kv[args[4]], nil
	}
	if len(args) >= 2 && args[0]=="status" { return "", nil }
	return strings.Join(args, " "), nil
}

func TestGitService_Concurrent_NoRaces(t *testing.T) {
	r1, r2 := &gRunner{}, &gRunner{}
	s1, s2 := NewService(r1), NewService(r2)
	var wg sync.WaitGroup
	wg.Add(2)
	go func(){ defer wg.Done(); _ = s1.Configure("Alice", "alice@example.com") }()
	go func(){ defer wg.Done(); _ = s2.ConfigureTx("Bob", "bob@example.com") }()
	wg.Wait()
}
