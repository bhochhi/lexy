# Lexy — Workflow-driven Chat Orchestrator (Go)

Lexy is a minimal, extensible chatbot service that executes JSON-defined workflows, calls NLU to detect intents, maintains per-client sessions, and dispatches named “codehooks” to intent/domain functions.

This repository favors:
- Clear boundaries via small interfaces and constructor injection
- Workflow-driven control flow with simple branching and templating
- A dispatcher pattern for calling hook functions by name from workflows

## Quick start

- Build: `go build ./...`
- Run server: (default 127.0.0.1:8080)
- Send a request:
  - POST /chat { "userText": "find my deductible", "clientId": "u1" }

NLU is mocked; workflows live under `workflow/`.

## Packages overview

- `main.go`
  - HTTP server and `/chat` endpoint; wires the orchestrator with concrete store, NLU client, and the composite hook registry.
- `orchestrator/`
  - Core workflow runner. Loads workflows, executes steps (validation | slot | message | fallback), handles retries, templating, and calls hooks via the Hook interface.
  - Depends on:
    - SessionStore (narrow interface) for state
    - NLU (interface) for DetectIntent
    - codehook.Hook to Invoke workflow functions
- `codehook/`
  - `hook.go`: neutral interface `Hook { Invoke(name string, ctx, args map[string]any) (map[string]any, error) }`
  - `domain/`: re-usable feature hooks, e.g., `banking`, `insurance` (namespaced calls: `banking.*`, `insurance.*`).
  - `intents/`: intent-specific hooks, e.g., `findDeductible`, `disputeTransaction`.
  - Pattern: each package exposes `New() Hook` and implements a single dispatcher `Invoke` that routes function names to internal handlers.
- `registry/`
  - Composite hook that routes Invoke calls to the right intent or domain hook.
  - Manual registration of intents/domains; returns a composite `Hook` used by the orchestrator.
- `nlu/`
  - `lex.go`: mocked NLU client implementing `DetectIntent` and returning slots.
- `session/`
  - In-memory session store (thread-safe) and session model with step states and context.
- `workflow/`
  - JSON workflow definitions; include both internal `intent` and business-facing `nluIntent` fields.

## Request flow (Mermaid)

The sequence below shows new vs existing sessions, NLU, workflow loading, and hook dispatch.

```mermaid
sequenceDiagram
    autonumber
    participant U as Client
    participant API as main.go (HTTP /chat)
    participant ORC as Orchestrator
    participant SS as SessionStore
    participant NLU as nlu.Client
    participant WF as Workflow Loader (FS)
    participant REG as registry.Composite (codehook.Hook)
    participant INT as Intent Hook (findDeductible)
    participant DOM as Domain Hook (insurance/banking)

    Note over API,REG: Startup: store := session.NewStore()<br/>funcs := registry.New()  ➜ composite Hook<br/>orc := orchestrator.New(store, nlu.NewClient(), funcs, workDir)

    U->>API: POST /chat { userText, clientId }
    API->>ORC: Execute(clientId, userText)

    ORC->>SS: Get(clientId)
    alt New session (not found)
        ORC->>NLU: DetectIntent(userText, clientId)
        NLU-->>ORC: nluIntent, slots
        ORC->>WF: loadWorkflowByNLU(nluIntent)
        WF-->>ORC: Workflow { intent, nluIntent, steps }
        Note right of ORC: Create session with ctx: {intent, nluIntent, clientId, slots...}<br/>CurrentStepID := first step
        ORC->>SS: Save(session)
    else Existing session
        ORC->>WF: loadWorkflow(ctx.nluIntent)
        WF-->>ORC: Workflow
    end

    loop Step processing (max 20 to avoid loops)
        ORC->>ORC: Find step by CurrentStepID

        alt Step.type == "validation" (pre)
            ORC->>REG: Invoke(pre.function, ctx, args)
            alt Domain-prefixed name (e.g., "insurance.hasInsuranceProduct")
                REG->>DOM: Invoke(fn, ctx, args)
                DOM-->>REG: result (e.g., {hasData: true})
            else Routed by ctx.intent
                REG->>INT: Invoke(fn, ctx, args)
                INT-->>REG: result (may mutate ctx)
            end
            REG-->>ORC: result
            ORC->>ORC: Branch via onSuccess/onFailure (may set next step)
        else Step.type == "slot"
            Note right of ORC: Build prompt (with retries) and collect options
            par For each options.functions[]
                ORC->>REG: Invoke(f.name, ctx, resolvedArgs)
                REG->>INT: Invoke(...)
                INT-->>REG: { options: [...] } (or [])
                REG-->>ORC: options
            and Static options
                ORC->>ORC: Append step.Options.static
            end
            opt Post-validation
                ORC->>REG: Invoke(post.function, ctx, {"value": capturedSlot})
                REG-->>ORC: result or error
                alt Resolver function in onSuccess (string, not a stepId)
                    ORC->>REG: Invoke(resolverFn, ctx, {"value": ...})
                    REG-->>ORC: {nextStep: "..."}
                end
                ORC->>ORC: Set next step or retry/failure
            end
            ORC->>SS: Save(session)  Note right of SS: Save after awaiting user input or transition
            ORC-->>API: Return prompt + options (when waiting for user input)
            API-->>U: 200 OK { responseText, options }
        else Step.type == "message" | "fallback"
            ORC->>ORC: Render contents with {{context.*}} templating
            ORC->>REG: collectOptions() via functions/static
            REG-->>ORC: options
            alt onSuccess == "end"
                ORC->>SS: Save(session)
                ORC-->>API: Return message + options
                API-->>U: 200 OK { responseText, options }
            else onSuccess is next step
                ORC->>ORC: Move to next step and continue
            end
        end
    end
```

The Mermaid source is also stored at `docs/sequence.mmd`.

## Adding a new intent

1) Create a new package under `codehook/intents/<intentName>` with:
   - `type hooks struct{ /* optional deps */ }`
   - `func New(/* deps */) codehook.Hook` returning `&hooks{...}`
   - `func (h *hooks) Invoke(name string, ctx, args map[string]any) (map[string]any, error)` that maps names to functions
2) Add a workflow JSON in `workflow/` with `intent` and (optionally) `nluIntent`.
3) Register it in `registry.New()`:
   - `RegisterIntent("<intentId>", intents.<intentName>.New(/* deps */))`

## Adding a new domain hook

1) Create a package under `codehook/domain/<domain>`; expose `New() codehook.Hook` and implement `Invoke`.
2) Use functions from workflows as `"<domain>.<function>"`.
3) Register it in `registry.New()` via `RegisterDomain("<domain>", domain.<domain>.New())`.

## Testing

- Invoke-based testing (unit):
  - `h := findDeductible.New()`
  - `res, err := h.Invoke("validatePolicySelection", ctx, map[string]any{"value": "Home Policy"})`
  - Assert on `res` and `ctx`.
- Orchestrator tests (integration-ish):
  - Pass a fake `codehook.Hook` that returns controlled outputs to drive branches.

## Design choices

- Dispatcher pattern (single `Invoke`) keeps the public surface small and matches name-based workflow calls.
- Neutral `codehook.Hook` interface decouples orchestrator from codehook implementations.
- Manual registration avoids init() import cycles and keeps wiring explicit.

---

Maintainers: update `registry.New()` for new hooks and `workflow/` for new flows. Keep functions small and pure where possible; inject upstream clients via constructors when needed.

## Codehook functions: roles and contracts

Codehook functions are intentionally small, composable building blocks that the workflow engine calls at specific stages. To keep workflow JSON minimal, each function follows a tiny contract based on its role. All functions share the same signature and return a plain map[string]any; the orchestrator interprets the result based on the call site.

- Signature (all hooks)
  - Invoke(name string, ctx map[string]any, args map[string]any) (map[string]any, error)
  - ctx: the per-session context (mutable by the function if appropriate)
  - args: function arguments resolved from workflow JSON (with {{context.*}} templating) and augmented by orchestrator when using NeedsSpec
  - return: a small result map, role-specific keys listed below; error indicates failure and triggers onFailure/retry per step

### Roles by step stage

1) Pre-validation (validation step, pre)
   - Purpose: probe upstream/state to decide early branching before user input
   - Inputs: ctx and args (often clientId/intent or service hints)
   - Output keys:
     - hasData: bool (true → onSuccess.hasData, false → onSuccess.hasNoData)
   - Errors: return error to route to onFailure

2) Dynamic options builder (slot/message steps)
   - Purpose: return options for the UI prompt
   - Output keys:
     - options: []session.Option or []map[string]any with {text, url}
   - Errors: propagate to let orchestrator choose onFailure for the step

3) Post-validation (validation step, post)
   - Purpose: validate user-provided slot value and enrich context
   - Inputs: args{"value": string}
   - Expected behavior: set context fields (e.g., policyType, policyNumber), return {ok: true}
   - Errors: non-nil error triggers retry (until maxRetries) then onFailure

4) Next-step resolver
   - Purpose: compute the next step id dynamically
   - Output keys:
     - nextStep: string (valid step id)
   - Errors: non-nil error routes to onFailure

5) Domain/API helpers (optional, reusable)
   - Purpose: encapsulate upstream calls shared across flows
   - Inputs: the orchestrator can provide HTTP invocation details via NeedsSpec; functions expect args to include method, uri, headers, body
   - Output: small decision result (e.g., {hasData: true}) or normalized payload for downstream steps

### Design guidelines

- Keep functions single-purpose and side-effect-light; set fields in ctx only when you need to share across subsequent steps.
- Prefer returning small, predictable shapes; avoid deeply nested structures.
- Log succinctly with slog attrs (logx):
  - logx.L().Info("validate_ok", slog.String("value", val))
  - logx.C(ctx).Error("upstream_failed", slog.Any("error", err))
- Inject external clients via constructors (registry.New wires them in) rather than global variables.

## Minimal JSON with NeedsSpec (optional, powerful)

To avoid bloating JSON with literal values, a function call can declare what it “needs.” The orchestrator resolves those into args before invocation.

- Where values come from:
  - ctx:key → from session context
  - args:key → from function args (after templating)
  - env:VAR → from process environment
  - gen:uuid → generated uuid
  - const:value → literal string

- Resolved keys added to args:
  - method: string
  - uri: string
  - headers: map[string]string (merged into args["headers"]) 
  - body: map[string]any (merged into args["body"]) 

### Example: dynamic options function

```
{
  "name": "banking.hasBankingProduct",
  "args": { "tenant": "acme" },
  "needs": {
    "method": "const:POST",
    "uri": "const:/post",
    "headers": {
      "Accept": "const:application/json",
      "Api-Key": "env:HTTPBIN_API_KEY",
      "X-Tenant": "args:tenant"
    },
    "body": {
      "client_id": "args:clientId",
      "query": "const:has_banking_product"
    }
  }
}
```

The hook can now read `args["method"], args["uri"], args["headers"], args["body"]` directly. If you prefer not to use NeedsSpec for a call, just supply `args` and keep the function logic simple.

### Example: post validation

```
"validation": {
  "post": {
    "function": "findDeductible.validatePolicySelection",
    "onSuccess": {
      "Auto Policy": "autoPolicyDetail",
      "Home Policy": "homePolicyDetail"
    }
  }
}
```

The function reads `args["value"]`, validates it, updates ctx (e.g., `ctx["policyType"] = val`), and returns `{"ok": true}`.

### Example: resolver function

```
{
  "name": "findDeductible.resolveNextStepFunction",
  "args": { "value": "{{context.policyType}}" }
}
```

The function returns `{"nextStep": "deductibleResult"}` (or computes based on args/ctx).

## When to choose NeedsSpec vs plain args

- Use plain args when values are already known (from templating or config) and the function is simple.
- Use NeedsSpec when the call needs a standard set of HTTP bits (method, uri, headers, body) sourced from context/env/args without hard-coding values in JSON.
- Keep JSON minimal: only pass overrides that are truly flow-specific (e.g., tenant); let the hook/spec fill the rest.
