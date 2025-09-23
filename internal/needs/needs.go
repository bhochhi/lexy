package needs

import (
    "os"
    "strings"
    "github.com/google/uuid"
)

// Ref describes where to get a value from.
type Ref struct {
    From    string // const|ctx|args|env|gen
    Key     string // key/path depending on From
    Optional bool
    Default any
}

// Spec declares required method, path, headers and body fields for a call.
type Spec struct {
    Method  string
    Path    string
    Headers map[string]Ref
    Body    map[string]Ref
}

// Helper constructors for concise specs
func Const(val any) Ref { return Ref{From: "const", Default: val} }
func Ctx(key string, optional bool) Ref { return Ref{From: "ctx", Key: key, Optional: optional} }
func Args(key string, optional bool) Ref { return Ref{From: "args", Key: key, Optional: optional} }
func Env(key string) Ref { return Ref{From: "env", Key: key} }
func Gen(kind string, optional bool) Ref { return Ref{From: "gen", Key: kind, Optional: optional} }

// Resolve produces concrete values by reading ctx, args, env, and generators.
func Resolve(spec Spec, ctx map[string]any, args map[string]any) (method string, path string, headers map[string]string, body map[string]any) {
    method = spec.Method
    if method == "" { method = "POST" }
    path = strings.TrimSpace(spec.Path)
    headers = make(map[string]string, len(spec.Headers))
    body = make(map[string]any, len(spec.Body))

    get := func(r Ref) (any, bool) {
        switch r.From {
        case "const":
            return r.Default, true
        case "ctx":
            if ctx == nil { return r.Default, false }
            v, ok := ctx[r.Key]
            if !ok || v == nil || v == "" { return r.Default, false }
            return v, true
        case "args":
            if args == nil { return r.Default, false }
            v, ok := args[r.Key]
            if !ok || v == nil || v == "" { return r.Default, false }
            return v, true
        case "env":
            if v := os.Getenv(r.Key); v != "" { return v, true }
            return r.Default, false
        case "gen":
            switch r.Key {
            case "uuid":
                u := uuid.NewString()
                return u, true
            }
            return r.Default, false
        default:
            return r.Default, false
        }
    }

    for k, r := range spec.Headers {
        if v, ok := get(r); ok {
            if s, ok2 := v.(string); ok2 { headers[k] = s }
        } else if !r.Optional {
            // keep absent; caller may validate if needed
        }
    }
    for k, r := range spec.Body {
        if v, ok := get(r); ok { body[k] = v }
    }
    return
}
 
