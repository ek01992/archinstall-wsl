package logx

import (
	"bytes"
	"log/slog"
	"testing"
)

func TestSetLoggerAndLoggingCalls(t *testing.T) {
	var buf bytes.Buffer
	prev := Logger
	h := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	l := slog.New(h)
	SetLogger(l)
	t.Cleanup(func() { SetLogger(prev) })

	Info("info test", "k", "v")
	Warn("warn test")
	Error("error test")

	if buf.Len() == 0 {
		t.Fatalf("expected logger to write something")
	}
}
