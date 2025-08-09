package firewall

import (
	"strings"
	"sync"
	"testing"
)

type cRunner struct{ out map[string]string }
func (c *cRunner) Run(name string, args ...string) error { return nil }
func (c *cRunner) Output(name string, args ...string) (string, error) { return c.out[name+" "+strings.Join(args, " ")], nil }

func TestFirewallService_Concurrent_NoRaces(t *testing.T) {
	fr1 := &cRunner{out: map[string]string{"ufw status": "Status: inactive"}}
	fr2 := &cRunner{out: map[string]string{"ufw status": "Status: active\nDefault: deny (incoming), allow (outgoing)"}}
	s1 := NewService(fr1)
	s2 := NewService(fr2)
	var wg sync.WaitGroup
	wg.Add(2)
	go func(){ defer wg.Done(); _ = s1.Configure() }()
	go func(){ defer wg.Done(); _ = s2.Configure() }()
	wg.Wait()
}
