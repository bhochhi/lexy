package logx

import (
	"context"
	"log/slog"
	"os"
	"runtime/debug"
	"time"
)

// Package logx now wraps Go's standard log/slog for simple, structured JSON logging.

var (
	defaultLogger *slog.Logger
	ctxLoggerKey  = &struct{}{}
)

// Init configures the default JSON logger with common attributes.
func Init(service string) *slog.Logger {
	level := slog.LevelInfo
	if os.Getenv("DEBUG") == "true" {
		level = slog.LevelDebug
	}
	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	defaultLogger = slog.New(h).With(
		slog.String("service", service),
		slog.String("env", getenv("ENV", "local")),
	)
	return defaultLogger
}

// Logger returns the default logger, initializing it if needed.
func Logger() *slog.Logger {
	if defaultLogger == nil {
		return Init("app")
	}
	return defaultLogger
}

// L returns the default logger (alias for Logger) for concise usage.
func L() *slog.Logger { return Logger() }

// From returns a logger from context or the default logger.
func From(ctx context.Context) *slog.Logger {
	if ctx == nil {
		return Logger()
	}
	if v := ctx.Value(ctxLoggerKey); v != nil {
		if l, ok := v.(*slog.Logger); ok {
			return l
		}
	}
	return Logger()
}

// C returns a logger derived from context (alias for From) for concise usage.
func C(ctx context.Context) *slog.Logger { return From(ctx) }

// WithLogger stores the provided logger into the context.
func WithLogger(ctx context.Context, l *slog.Logger) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, ctxLoggerKey, l)
}

// WithFields adds fields to the logger stored in the context and returns the new context.
func WithFields(ctx context.Context, fields map[string]any) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	base := From(ctx).With(toAttrs(fields)...)
	return WithLogger(ctx, base)
}

// WithAttrs adds slog.Attr fields to the logger stored in the context.
func WithAttrs(ctx context.Context, attrs ...slog.Attr) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	as := make([]any, 0, len(attrs))
	for _, a := range attrs {
		as = append(as, a)
	}
	base := From(ctx).With(as...)
	return WithLogger(ctx, base)
}

// Convenience helpers without context
func Debug(msg string, fields map[string]any) { Logger().Debug(msg, toAttrs(fields)...) }
func Info(msg string, fields map[string]any)  { Logger().Info(msg, toAttrs(fields)...) }
func Warn(msg string, fields map[string]any)  { Logger().Warn(msg, toAttrs(fields)...) }
func Error(msg string, fields map[string]any) { Logger().Error(msg, toAttrs(fields)...) }

// Context-aware helpers
func DebugContext(ctx context.Context, msg string, fields map[string]any) {
	From(ctx).Debug(msg, toAttrs(fields)...)
}
func InfoContext(ctx context.Context, msg string, fields map[string]any) {
	From(ctx).Info(msg, toAttrs(fields)...)
}
func WarnContext(ctx context.Context, msg string, fields map[string]any) {
	From(ctx).Warn(msg, toAttrs(fields)...)
}
func ErrorContext(ctx context.Context, msg string, fields map[string]any) {
	From(ctx).Error(msg, toAttrs(fields)...)
}

// Helper to format errors consistently
func ErrField(err error) map[string]any {
	if err == nil {
		return nil
	}
	return map[string]any{"error": err.Error()}
}

// ReqFields is a convenience to build common HTTP request/response fields
func ReqFields(status int, dur time.Duration, size int64) map[string]any {
	return map[string]any{
		"status":      status,
		"duration_ms": float64(dur.Microseconds()) / 1000.0,
		"size":        size,
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func toAttrs(fields map[string]any) []any {
	if len(fields) == 0 {
		return nil
	}
	attrs := make([]any, 0, len(fields))
	for k, v := range fields {
		attrs = append(attrs, slog.Any(k, v))
	}
	return attrs
}

// --- Ergonomic KV helpers ---------------------------------------------------

// InfoKV et al. allow writing fields as variadic key/value pairs instead of a map.
func DebugKV(msg string, kv ...any) { Logger().Debug(msg, kv...) }
func InfoKV(msg string, kv ...any)  { Logger().Info(msg, kv...) }
func WarnKV(msg string, kv ...any)  { Logger().Warn(msg, kv...) }
func ErrorKV(msg string, kv ...any) { Logger().Error(msg, kv...) }

func DebugKVContext(ctx context.Context, msg string, kv ...any) { From(ctx).Debug(msg, kv...) }
func InfoKVContext(ctx context.Context, msg string, kv ...any)  { From(ctx).Info(msg, kv...) }
func WarnKVContext(ctx context.Context, msg string, kv ...any)  { From(ctx).Warn(msg, kv...) }
func ErrorKVContext(ctx context.Context, msg string, kv ...any) { From(ctx).Error(msg, kv...) }

// ErrorWith includes error and optional stack trace (when DEBUG=true or LOG_STACK=1).
func ErrorWith(msg string, err error, kv ...any) {
	if err != nil {
		kv = append(kv, "error", err.Error())
		if os.Getenv("DEBUG") == "true" || os.Getenv("LOG_STACK") == "1" {
			kv = append(kv, "stack", string(debug.Stack()))
		}
	}
	Logger().Error(msg, kv...)
}

func ErrorWithContext(ctx context.Context, msg string, err error, kv ...any) {
	if err != nil {
		kv = append(kv, "error", err.Error())
		if os.Getenv("DEBUG") == "true" || os.Getenv("LOG_STACK") == "1" {
			kv = append(kv, "stack", string(debug.Stack()))
		}
	}
	From(ctx).Error(msg, kv...)
}
