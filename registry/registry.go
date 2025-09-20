package registry

import (
	"strings"
	"sync"

	"github.com/bhochhi/lexy/orchestrator"

	// Bring codehook packages for types; registration is done below explicitly.
	chBanking "github.com/bhochhi/lexy/codehook/domain/banking"
	chInsurance "github.com/bhochhi/lexy/codehook/domain/insurance"
	chDispute "github.com/bhochhi/lexy/codehook/intents/disputeTransaction"
	chDeductible "github.com/bhochhi/lexy/codehook/intents/findDeductible"
)

// New returns a single IntentFunctions wrapping whatever intents self-registered.
func New() orchestrator.IntentFunctions {
	// Manual registration to avoid init() and import cycles.
	RegisterDomain("banking", &chBanking.Registry{})
	RegisterDomain("insurance", &chInsurance.Registry{})
	RegisterIntent("disputeTransaction", chDispute.NewRegistry())
	RegisterIntent("findDeductible", chDeductible.NewRegistry())

	return newCompositeFuncs(All())
}

// Composite forwards function calls to the registry matching the current session intent.
// It allows adding new intent packages without changing orchestrator or main logic.
type composite struct {
	byIntent map[string]orchestrator.IntentFunctions
}

func newCompositeFuncs(byIntent map[string]orchestrator.IntentFunctions) *composite {
	return &composite{byIntent: byIntent}
}

func (c *composite) Call(name string, ctx map[string]any, args map[string]any) (map[string]any, error) {
	if ctx == nil {
		ctx = map[string]any{}
	}
	// Support namespaced function calls: "banking.hasCreditCards" or "insurance.hasInsuranceProduct"
	if i := strings.IndexByte(name, '.'); i > 0 {
		domain := name[:i]
		fn := name[i+1:]
		if reg, ok := AllDomains()[domain]; ok {
			return reg.Call(fn, ctx, args)
		}
	}
	intent, _ := ctx["intent"].(string)
	if reg, ok := c.byIntent[intent]; ok {
		return reg.Call(name, ctx, args)
	}
	// If no active intent found, try a default intent if provided
	if reg, ok := c.byIntent["default"]; ok {
		return reg.Call(name, ctx, args)
	}
	// Return empty map (no-op) to keep flow from crashing; errors will route to onFailure in orchestrator if used.
	return map[string]any{}, nil
}

//register

var (
	mu      sync.RWMutex
	intents = map[string]orchestrator.IntentFunctions{}
	domMu   sync.RWMutex
	domains = map[string]orchestrator.IntentFunctions{}
)

// RegisterIntent registers an intent-scoped function registry by name.
func RegisterIntent(intent string, r orchestrator.IntentFunctions) {
	mu.Lock()
	intents[intent] = r
	mu.Unlock()
}

// RegisterDomain registers a domain-scoped function registry (e.g., banking, insurance).
func RegisterDomain(domain string, r orchestrator.IntentFunctions) {
	domMu.Lock()
	domains[domain] = r
	domMu.Unlock()
}

// All returns a shallow copy of registered intent registries.
func All() map[string]orchestrator.IntentFunctions {
	mu.RLock()
	defer mu.RUnlock()
	out := make(map[string]orchestrator.IntentFunctions, len(intents))
	for k, v := range intents {
		out[k] = v
	}
	return out
}

// AllDomains returns a shallow copy of registered domain registries.
func AllDomains() map[string]orchestrator.IntentFunctions {
	domMu.RLock()
	defer domMu.RUnlock()
	out := make(map[string]orchestrator.IntentFunctions, len(domains))
	for k, v := range domains {
		out[k] = v
	}
	return out
}
