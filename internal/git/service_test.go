package git

import (
	"errors"
	"strings"
	"testing"
)

type fakeRunner struct{ out map[string]string }
func (f *fakeRunner) Run(name string, args ...string) error { return nil }
func (f *fakeRunner) Output(name string, args ...string) (string, error) {
	key := name + " " + strings.Join(args, " ")
	v, ok := f.out[key]
	if !ok {
		return "", errors.New("no output")
	}
	return v, nil
}

func TestService_Configure_SetsAndVerifies(t *testing.T) {
	fr := &fakeRunner{out: map[string]string{
		"git config --global --get user.name":  "Alice\n",
		"git config --global --get user.email": "alice@example.com\n",
	}}
	s := NewService(fr)
	if err := s.Configure("Alice", "alice@example.com"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
