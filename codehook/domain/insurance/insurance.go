package insurance

import (
	"strings"

	"github.com/bhochhi/lexy/orchestrator"
)

// Registry of insurance domain functions

type Registry struct{}

func (r *Registry) Call(name string, ctx map[string]any, args map[string]any) (map[string]any, error) {
	switch name {
	case "hasInsuranceProduct":
		return hasInsuranceProduct(ctx, args)
	default:
		return map[string]any{}, nil
	}
}

// Registration is manual via registry.New(); no init-based registration here.

// --- Domain funcs ---

func hasInsuranceProduct(ctx map[string]any, args map[string]any) (map[string]any, error) {
	// Mock: insurance product if clientId contains 'ins'
	cid, _ := ctx["clientId"].(string)
	return map[string]any{"hasData": strings.Contains(cid, "ins")}, nil
}

var _ orchestrator.IntentFunctions = (*Registry)(nil)
