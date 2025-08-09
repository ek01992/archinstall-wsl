package golang

import (
	"strings"
	"testing"
)

type fr struct{ out map[string]string }

func (f *fr) Run(name string, args ...string) error { return nil }
func (f *fr) Output(name string, args ...string) (string, error) {
	return f.out[name+" "+strings.Join(args, " ")], nil
}

type fv struct {
	v   string
	err error
}

func (fv fv) LatestGo() (string, error) { return fv.v, fv.err }

func TestService_Install_Idempotent(t *testing.T) {
	r := &fr{out: map[string]string{"go version": "go version go1.20.0 linux/amd64"}}
	s := NewService(r, fv{v: "1.20.0"})
	if err := s.Install(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestService_Install_Update_VerifyMismatch(t *testing.T) {
	r := &fr{out: map[string]string{"go version": "go version go1.19.0 linux/amd64"}}
	s := NewService(r, fv{v: "1.20.0"})
	if err := s.Install(); err == nil {
		t.Fatalf("expected verification failure when version mismatch remains")
	}
}
