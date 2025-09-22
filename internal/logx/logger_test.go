package logx

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// helper to create a buffer-backed JSON logger at a given level
func newBufferLogger(level slog.Level) (*slog.Logger, *bytes.Buffer) {
	var buf bytes.Buffer
	h := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: level})
	l := slog.New(h).With(slog.String("service", "test"), slog.String("env", "test"))
	return l, &buf
}

func decodeLines(buf *bytes.Buffer) []map[string]any {
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	var out []map[string]any
	for _, ln := range lines {
		if strings.TrimSpace(ln) == "" {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(ln), &m); err == nil {
			out = append(out, m)
		}
	}
	return out
}

func TestInfoWritesJSON(t *testing.T) {
	l, buf := newBufferLogger(slog.LevelInfo)
	defaultLogger = l
	Info("hello", map[string]any{"a": 1, "b": "x"})
	recs := decodeLines(buf)
	if len(recs) != 1 {
		t.Fatalf("expected 1 record, got %d", len(recs))
	}
	if recs[0]["msg"] != "hello" {
		t.Fatalf("unexpected msg: %v", recs[0]["msg"])
	}
	if recs[0]["a"].(float64) != 1 {
		t.Fatalf("missing/invalid field a: %v", recs[0]["a"])
	}
	if recs[0]["b"].(string) != "x" {
		t.Fatalf("missing/invalid field b: %v", recs[0]["b"])
	}
}

func TestDebugFilteredAtInfoLevel(t *testing.T) {
	l, buf := newBufferLogger(slog.LevelInfo)
	defaultLogger = l
	Debug("nope", nil)
	if buf.Len() != 0 {
		t.Fatalf("expected no output at info level, got: %s", buf.String())
	}
}

func TestInfoContextWithFields(t *testing.T) {
	l, buf := newBufferLogger(slog.LevelInfo)
	defaultLogger = l
	ctx := context.Background()
	ctx = WithFields(ctx, map[string]any{"request_id": "r1"})
	InfoContext(ctx, "ctx_msg", map[string]any{"k": "v"})
	recs := decodeLines(buf)
	if len(recs) != 1 {
		t.Fatalf("expected 1 record, got %d", len(recs))
	}
	if recs[0]["msg"] != "ctx_msg" {
		t.Fatalf("unexpected msg: %v", recs[0]["msg"])
	}
	if recs[0]["request_id"].(string) != "r1" {
		t.Fatalf("missing request_id: %v", recs[0]["request_id"])
	}
	if recs[0]["k"].(string) != "v" {
		t.Fatalf("missing k: %v", recs[0]["k"])
	}
}

func TestLoggingMiddlewareEmitsStartAndEnd(t *testing.T) {
	l, buf := newBufferLogger(slog.LevelInfo)
	defaultLogger = l

	base := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	h := LoggingMiddleware(base)

	req := httptest.NewRequest(http.MethodPost, "/chat", strings.NewReader("{}"))
	req.Header.Set(HRequestID, "req-123")
	req.Header.Set(HClientID, "c-42")
	req.Header.Set(HConversationID, "conv-9")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	recs := decodeLines(buf)
	if len(recs) != 2 {
		t.Fatalf("expected 2 records, got %d (%v)", len(recs), recs)
	}
	if recs[0]["msg"] != "http_request_start" {
		t.Fatalf("unexpected first msg: %v", recs[0]["msg"])
	}
	if recs[0]["method"].(string) != http.MethodPost {
		t.Fatalf("missing method: %v", recs[0]["method"])
	}
	if recs[0]["path"].(string) != "/chat" {
		t.Fatalf("missing path: %v", recs[0]["path"])
	}
	if recs[0]["request_id"].(string) != "req-123" {
		t.Fatalf("missing request_id: %v", recs[0]["request_id"])
	}
	if recs[1]["msg"] != "http_request_end" {
		t.Fatalf("unexpected second msg: %v", recs[1]["msg"])
	}
	if recs[1]["status"].(float64) != 200 {
		t.Fatalf("missing status: %v", recs[1]["status"])
	}
}

func TestLogErrorHelper(t *testing.T) {
	l, buf := newBufferLogger(slog.LevelInfo)
	defaultLogger = l
	ctx := context.Background()
	LogError(ctx, "oops", errors.New("bad"))
	recs := decodeLines(buf)
	if len(recs) != 1 {
		t.Fatalf("expected 1 record, got %d", len(recs))
	}
	if recs[0]["msg"] != "oops" {
		t.Fatalf("unexpected msg: %v", recs[0]["msg"])
	}
	if recs[0]["error"].(string) != "bad" {
		t.Fatalf("missing error field: %v", recs[0]["error"])
	}
}
