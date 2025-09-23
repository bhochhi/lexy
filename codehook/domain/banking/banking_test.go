package banking

import (
    "context"
    "net/http"
    "testing"
)

type mockHTTP struct{
    status int
}

func (m mockHTTP) PostJSON(ctx context.Context, path string, headers map[string]string, body any) (map[string]any, int, error) {
    return map[string]any{"ok": true}, m.status, nil
}

func TestHasBankingProduct_UsesHTTPClientAndCtx(t *testing.T) {
    h := &hooks{ http: mockHTTP{status: http.StatusOK} }
    ctx := map[string]any{"clientId": "acme-bank-1"}
    out, err := h.hasBankingProduct(ctx, nil)
    if err != nil { t.Fatalf("unexpected err: %v", err) }
    if v, _ := out["hasData"].(bool); !v { t.Fatalf("expected hasData true, got %v", out) }
}
