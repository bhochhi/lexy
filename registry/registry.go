package registry

import (
	"os"
	"strings"
	"sync"

	"github.com/bhochhi/lexy/codehook"

	// Bring codehook packages for types; registration is done below explicitly.
	chBanking "github.com/bhochhi/lexy/codehook/domain/banking"
	chInsurance "github.com/bhochhi/lexy/codehook/domain/insurance"
	chDispute "github.com/bhochhi/lexy/codehook/intents/disputeTransaction"
	chDeductible "github.com/bhochhi/lexy/codehook/intents/findDeductible"
	"github.com/bhochhi/lexy/internal/httpc"
)

// New returns a single IntentFunctions wrapping whatever intents self-registered.
func New() codehook.Hook {
	// Construct shared clients/deps
	httpBase := getenv("HTTPBIN_URL", "https://httpbin.org")
	httpKey := getenv("HTTPBIN_API_KEY", "")
	httpClient := httpc.NewHTTPBin(httpBase, httpKey)

	// Manual registration to avoid init() and import cycles.
	RegisterDomain("banking", chBanking.NewWithHTTP(httpClient))
	RegisterDomain("insurance", chInsurance.New())
	RegisterIntent("disputeTransaction", chDispute.New())
	RegisterIntent("findDeductible", chDeductible.New())

	return newCompositeFuncs(All())
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" { return v }
	return def
}

// Composite forwards function calls to the registry matching the current session intent.
// It allows adding new intent packages without changing orchestrator or main logic.
type composite struct {
	byIntent map[string]codehook.Hook
}

func newCompositeFuncs(byIntent map[string]codehook.Hook) *composite {
	return &composite{byIntent: byIntent}
}

func (c *composite) Invoke(name string, ctx map[string]any, args map[string]any) (map[string]any, error) {
	if ctx == nil {
		ctx = map[string]any{}
	}
	// Support namespaced function calls: "banking.hasCreditCards" or "insurance.hasInsuranceProduct"
	if i := strings.IndexByte(name, '.'); i > 0 {
		domain := name[:i]
		fn := name[i+1:]
		if reg, ok := AllDomains()[domain]; ok {
			return reg.Invoke(fn, ctx, args)
		}
	}
	intent, _ := ctx["intent"].(string)
	if reg, ok := c.byIntent[intent]; ok {
		return reg.Invoke(name, ctx, args)
	}
	// If no active intent found, try a default intent if provided
	if reg, ok := c.byIntent["default"]; ok {
		return reg.Invoke(name, ctx, args)
	}
	// Return empty map (no-op) to keep flow from crashing; errors will route to onFailure in orchestrator if used.
	return map[string]any{}, nil
}

//register

var (
	mu      sync.RWMutex
	intents = map[string]codehook.Hook{}
	domMu   sync.RWMutex
	domains = map[string]codehook.Hook{}
)

// RegisterIntent registers an intent-scoped function registry by name.
func RegisterIntent(intent string, r codehook.Hook) {
	mu.Lock()
	intents[intent] = r
	mu.Unlock()
}

// RegisterDomain registers a domain-scoped function registry (e.g., banking, insurance).
func RegisterDomain(domain string, r codehook.Hook) {
	domMu.Lock()
	domains[domain] = r
	domMu.Unlock()
}

// All returns a shallow copy of registered intent registries.
func All() map[string]codehook.Hook {
	mu.RLock()
	defer mu.RUnlock()
	out := make(map[string]codehook.Hook, len(intents))
	for k, v := range intents {
		out[k] = v
	}
	return out
}

// AllDomains returns a shallow copy of registered domain registries.
func AllDomains() map[string]codehook.Hook {
	domMu.RLock()
	defer domMu.RUnlock()
	out := make(map[string]codehook.Hook, len(domains))
	for k, v := range domains {
		out[k] = v
	}
	return out
}
