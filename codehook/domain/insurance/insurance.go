package insurance

import (
	"strings"

	"github.com/bhochhi/lexy/codehook"
)

// hooks implements the insurance domain functions dispatcher.
type hooks struct{}

// New returns the codehook.Hook for the insurance domain.
func New() codehook.Hook { return &hooks{} }

func (r *hooks) Invoke(name string, ctx map[string]any, args map[string]any) (map[string]any, error) {
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

// Compile-time check (optional)
var _ codehook.Hook = (*hooks)(nil)
