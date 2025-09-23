package httpc

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "strings"
    "time"
)

// HTTPBinClient defines the minimal capability we need from an HTTP client.
type HTTPBinClient interface {
    PostJSON(ctx context.Context, path string, headers map[string]string, body any) (map[string]any, int, error)
}

// HTTPBin implements HTTPBinClient against a JSON endpoint.
type HTTPBin struct {
    baseURL string
    apiKey  string
    hc      *http.Client
}

func NewHTTPBin(baseURL, apiKey string) *HTTPBin {
    if !strings.HasPrefix(baseURL, "http") {
        baseURL = "https://" + baseURL
    }
    return &HTTPBin{
        baseURL: strings.TrimRight(baseURL, "/"),
        apiKey:  apiKey,
        hc:      &http.Client{Timeout: 10 * time.Second},
    }
}

func (c *HTTPBin) PostJSON(ctx context.Context, path string, headers map[string]string, body any) (map[string]any, int, error) {
    b, err := json.Marshal(body)
    if err != nil { return nil, 0, fmt.Errorf("marshal body: %w", err) }
    url := c.baseURL + "/" + strings.TrimLeft(path, "/")
    req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
    if err != nil { return nil, 0, fmt.Errorf("new request: %w", err) }
    // default headers
    req.Header.Set("Accept", "application/json")
    req.Header.Set("Content-Type", "application/json")
    if c.apiKey != "" { req.Header.Set("Api-Key", c.apiKey) }
    // merge custom headers
    for k, v := range headers { req.Header.Set(k, v) }

    resp, err := c.hc.Do(req)
    if err != nil { return nil, 0, fmt.Errorf("do request: %w", err) }
    defer resp.Body.Close()
    rb, err := io.ReadAll(resp.Body)
    if err != nil { return nil, resp.StatusCode, fmt.Errorf("read body: %w", err) }
    var m map[string]any
    if err := json.Unmarshal(rb, &m); err != nil {
        // return raw text if not JSON
        m = map[string]any{"raw": string(rb)}
    }
    return m, resp.StatusCode, nil
}
