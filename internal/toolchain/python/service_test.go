package python

import (
	"strings"
	"testing"
)

type pr struct{ out map[string]string }
func (f *pr) Run(name string, args ...string) error { return nil }
func (f *pr) Output(name string, args ...string) (string, error) { return f.out[name+" "+strings.Join(args, " ")], nil }

type pv struct{ v string; err error }
func (pv pv) LatestPython() (string, error) { return pv.v, pv.err }

func TestService_Install_Idempotent(t *testing.T) {
	r := &pr{out: map[string]string{"python --version": "Python 3.12.4", "pipx --version": "1.4.5"}}
	s := NewService(r, pv{v: "3.12.4"})
	if err := s.Install(); err != nil { t.Fatalf("unexpected error: %v", err) }
}
