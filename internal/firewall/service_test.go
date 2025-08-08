package firewall

import (
	"strings"
	"testing"
)

type fakeRunner struct{ out map[string]string; last []string }
func (f *fakeRunner) Run(name string, args ...string) error { f.last = append([]string{name}, args...); return nil }
func (f *fakeRunner) Output(name string, args ...string) (string, error) { return f.out[name+" "+strings.Join(args, " ")], nil }

func TestService_Configure_IdempotentActive(t *testing.T) {
	fr := &fakeRunner{out: map[string]string{
		"ufw status": "Status: active\nDefault: deny (incoming), allow (outgoing)\nAnywhere ALLOW 172.16.0.0/12\n",
	}}
	s := NewService(fr)
	if err := s.Configure(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
