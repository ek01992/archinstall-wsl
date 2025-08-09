package nodejs

import (
	"errors"
	"strings"
	"testing"
)

type nr struct{ out map[string]string }
func (f *nr) Run(name string, args ...string) error { return nil }
func (f *nr) Output(name string, args ...string) (string, error) { return f.out[name+" "+strings.Join(args, " ")] , nil }
func (f *nr) Shell(cmd string) (string, error) { if cmd == "node -v" { return f.out["node -v"], nil }; return "", errors.New("no shell") }

type nv struct{ v string; err error }
func (nv nv) LatestLTS() (string, error) { return nv.v, nv.err }

func TestService_Install_Idempotent(t *testing.T) {
	r := &nr{out: map[string]string{"nvm --version": "0.39.7", "node -v": "v20.16.0"}}
	s := NewService(r, nv{v: "v20.16.0"})
	if err := s.Install(); err != nil { t.Fatalf("unexpected error: %v", err) }
}

func TestService_Install_InstallAndAlias(t *testing.T) {
	r := &nr{out: map[string]string{"nvm --version": "0.39.7", "node -v": "v18.0.0"}}
	s := NewService(r, nv{v: "v20.0.0"})
	if err := s.Install(); err == nil {
		// verification should fail because current stays at v18.0.0
		t.Fatalf("expected verification failure when version mismatch")
	}
}
