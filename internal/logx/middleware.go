package logx

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

// Common header names (case-insensitive by Go's http) we care about
const (
	HClientID       = "Client-Id"
	HConversationID = "Conversation-Id"
	HRequestID      = "X-Request-Id"
)

// LoggingMiddleware wraps an http.Handler to add request-scoped context and structured logs.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		// Build a request-scoped logger with common attrs and attach to context
		logger := Logger().With(
			slog.String("request_id", r.Header.Get(HRequestID)),
			slog.String("client_id", r.Header.Get(HClientID)),
			slog.String("conversation_id", r.Header.Get(HConversationID)),
			slog.String("path", r.URL.Path),
			slog.String("method", r.Method),
		)
		ctx := WithLogger(r.Context(), logger)

		// log request start
		InfoContext(ctx, "http_request_start", map[string]any{
			"content_length": r.ContentLength,
		})

		// wrap ResponseWriter to capture status and size
		rw := &respWriter{ResponseWriter: w, status: 200}
		next.ServeHTTP(rw, r.WithContext(ctx))

		dur := time.Since(start)
		InfoContext(ctx, "http_request_end", ReqFields(rw.status, dur, rw.size))
	})
}

type respWriter struct {
	http.ResponseWriter
	status int
	size   int64
}

func (w *respWriter) WriteHeader(statusCode int) {
	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *respWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.size += int64(n)
	return n, err
}

// Optional helper to log errors during handler execution
func LogError(ctx context.Context, msg string, err error) {
	ErrorContext(ctx, msg, ErrField(err))
}

// Helper to safely copy request body if needed
func readAllLimit(r io.Reader, limit int64) ([]byte, error) {
	lr := &io.LimitedReader{R: r, N: limit}
	return io.ReadAll(lr)
}

// ParseInt helper for query/headers
func parseInt(s string, def int) int {
	if s == "" {
		return def
	}
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	return def
}
