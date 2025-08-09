package firewall

import (
	"strings"
	"testing"
)

type txRunner struct{ out map[string]string }
func (t *txRunner) Run(name string, args ...string) error { return nil }
func (t *txRunner) Output(name string, args ...string) (string, error) { return t.out[name+" "+strings.Join(args, " ")], nil }

func TestService_ConfigureTx_Active_NoChange(t *testing.T) {
	r := &txRunner{out: map[string]string{"ufw status": "Status: active\nDefault: deny (incoming), allow (outgoing)\nAnywhere ALLOW 172.16.0.0/12\n"}}
	s := NewService(r)
	if err := s.ConfigureTx(); err != nil { t.Fatalf("unexpected error: %v", err) }
}
