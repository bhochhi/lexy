package banking

import (
	"context"
	"strings"

	"github.com/bhochhi/lexy/codehook"
	"github.com/bhochhi/lexy/internal/httpc"
	"github.com/bhochhi/lexy/internal/logx"
	"log/slog"
	"os"
)

// hooks implements the banking domain functions dispatcher.
type hooks struct{
	http HTTPClient
}

// HTTPClient is the minimal interface this package depends on; easy to mock in tests.
type HTTPClient interface {
	PostJSON(ctx context.Context, path string, headers map[string]string, body any) (map[string]any, int, error)
}

// New returns the codehook.Hook for the banking domain.
func New() codehook.Hook {
	// Defaults can be overridden later if we add a DI registry; for now, build from env.
	base := getenv("HTTPBIN_URL", "https://httpbin.org")
	key := getenv("HTTPBIN_API_KEY", "")
	return &hooks{ http: httpc.NewHTTPBin(base, key) }
}

// NewWithHTTP returns the codehook.Hook using the provided HTTP client (for DI/tests).
func NewWithHTTP(c HTTPClient) codehook.Hook {
	return &hooks{ http: c }
}

func (r *hooks) Invoke(name string, ctx map[string]any, args map[string]any) (map[string]any, error) {
	switch name {
	case "hasBankingProduct":
		return r.hasBankingProduct(ctx, args)
	case "hasCreditCards":
		return hasCreditCards(ctx, args)
	default:
		return map[string]any{}, nil
	}
}

// Registration is manual via registry.New(); no init-based registration here.

// --- Domain funcs ---

func (r *hooks) hasBankingProduct(ctx map[string]any, args map[string]any) (map[string]any, error) {
	// Expect method/uri/headers/body resolved by orchestrator via FunctionSpec.Needs
	// with reasonable defaults here.
	method, _ := args["method"].(string)
	if method == "" { method = "POST" }
	path, _ := args["uri"].(string)
	if path == "" { path = "/post" }
	// headers can be map[string]any or map[string]string
	hdrs := map[string]string{"Accept": "application/json"}
	if hAny, ok := args["headers"].(map[string]any); ok {
		for k, v := range hAny { hdrs[k] = toString(v) }
	} else if hs, ok := args["headers"].(map[string]string); ok {
		for k, v := range hs { hdrs[k] = v }
	}
	body, _ := args["body"].(map[string]any)
	if body == nil { body = map[string]any{} }
	// Ensure client_id exists in body; fallback from session ctx
	if _, ok := body["client_id"]; !ok {
		if v, _ := ctx["clientId"].(string); v != "" { body["client_id"] = v }
	}
	cid, _ := body["client_id"].(string)
	if cid == "" { return map[string]any{"hasData": false}, nil }

	// Use request-scoped logger if available
	c := logx.WithAttrs(context.Background(), slog.String("function", "banking.hasBankingProduct"))
	// Call external service
	_ = method // reserved for future when methods vary
	resp, status, err := r.http.PostJSON(c, path, hdrs, body)
	if err != nil {
		logx.ErrorWithContext(c, "httpbin_call_failed", err, "status", status)
		return map[string]any{"hasData": false, "error": "remote_error"}, nil
	}
	logx.C(c).Debug("httpbin_response", slog.Int("status", status), slog.Any("resp", resp))

	// Example decision: treat 200 as positive and echo our mock condition
	has := status == 200 && strings.Contains(cid, "bank")
	return map[string]any{"hasData": has}, nil
}

func hasCreditCards(ctx map[string]any, args map[string]any) (map[string]any, error) {
	// Mock: if clientId ends with an even digit
	cid, _ := ctx["clientId"].(string)
	if cid == "" {
		return map[string]any{"hasData": false}, nil
	}
	last := cid[len(cid)-1]
	return map[string]any{"hasData": last%2 == 0}, nil
}

// Compile-time check (optional)
var _ codehook.Hook = (*hooks)(nil)

// getenv local helper
func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" { return v }
	return def
}

func toString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	default:
		return ""
	}
}
