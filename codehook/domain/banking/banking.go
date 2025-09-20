package banking

import (
	"strings"

	"github.com/bhochhi/lexy/orchestrator"
)

// Registry of banking domain functions

type Registry struct{}

func (r *Registry) Call(name string, ctx map[string]any, args map[string]any) (map[string]any, error) {
	switch name {
	case "hasBankingProduct":
		return hasBankingProduct(ctx, args)
	case "hasCreditCards":
		return hasCreditCards(ctx, args)
	default:
		return map[string]any{}, nil
	}
}

// Registration is manual via registry.New(); no init-based registration here.

// --- Domain funcs ---

func hasBankingProduct(ctx map[string]any, args map[string]any) (map[string]any, error) {
	// Mock: banking product if clientId contains 'bank'
	cid, _ := ctx["clientId"].(string)
	return map[string]any{"hasData": strings.Contains(cid, "bank")}, nil
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

var _ orchestrator.IntentFunctions = (*Registry)(nil)
