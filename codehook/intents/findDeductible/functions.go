package findDeductible

import (
	"errors"
	"fmt"
	"strings"

	"github.com/bhochhi/lexy/session"
)

// Registry implements orchestrator.IntentFunctions for this intent.
type Registry struct{}

func NewRegistry() *Registry { return &Registry{} }

// Call dispatches function by name.
func (r *Registry) Call(name string, ctx map[string]any, args map[string]any) (map[string]any, error) {
	funcMap := map[string]func(map[string]any, map[string]any) (map[string]any, error){
		"checkClientPolicies":      checkClientPolicies,
		"getDynamicPolicies":       getDynamicPolicies,
		"validatePolicySelection":  validatePolicySelection,
		"getClientVehicles":        getClientVehicles,
		"validateVehicleSelection": validateVehicleInput,
		"getHomePolicies":          retrieveHomePolicyOptions,
		"validateHomeSelection":    validateHomeSelection,
		"getRenterPolicies":        getRenterPolicies,
		"validateRenterSelection":  validateRenterSelection,
		"getPolicyPortalLink":      getPolicyPortalLink,
		"getContactOptions":        getContactOptions,
		"resolveNextStepFunction":  resolveNextStep,
	}
	if fn, ok := funcMap[name]; ok {
		return fn(ctx, args)
	}
	return nil, fmt.Errorf("unknown function: %s", name)
}

// Registration is manual via registry.New(); no init-based registration here.

// --- Function implementations (mocked) ---

func checkClientPolicies(ctx map[string]any, args map[string]any) (map[string]any, error) {
	// Pretend to check via API; here we mock that client has data unless clientId ends with 0
	cid, _ := args["clientId"].(string)
	if cid == "" {
		cid, _ = ctx["clientId"].(string)
	}
	if strings.HasSuffix(cid, "0") {
		return map[string]any{"hasData": false, "hasNoData": true}, nil
	}
	return map[string]any{"hasData": true, "hasNoData": false}, nil
}

func getDynamicPolicies(ctx map[string]any, args map[string]any) (map[string]any, error) {
	// Build options dynamically; in real code, query user policies
	options := []session.Option{
		{Text: "Umbrella Policy"},
	}
	return map[string]any{"options": options}, nil
}

func validatePolicySelection(ctx map[string]any, args map[string]any) (map[string]any, error) {
	val, _ := args["value"].(string)
	val = strings.TrimSpace(val)
	if val == "" {
		return nil, errors.New("empty selection")
	}
	// accept if known
	switch val {
	case "Auto Policy", "Home Policy", "Renter Policy", "Umbrella Policy":
		ctx["policyType"] = val
		return map[string]any{"ok": true}, nil
	default:
		return nil, errors.New("invalid policy type")
	}
}

func getClientVehicles(ctx map[string]any, args map[string]any) (map[string]any, error) {
	// Mock two vehicles
	options := []session.Option{
		{Text: "2019 Honda Civic - ABC123"},
		{Text: "2021 Tesla Model 3 - XYZ789"},
	}
	return map[string]any{"options": options}, nil
}

func validateVehicleInput(ctx map[string]any, args map[string]any) (map[string]any, error) {
	val, _ := args["value"].(string)
	if val == "" {
		return nil, errors.New("no vehicle")
	}
	parts := strings.Split(val, " ")
	// Expected: "2019 Honda Civic - ABC123" -> policy number token likely last item
	if len(parts) == 0 {
		return nil, errors.New("invalid vehicle format")
	}
	token := parts[len(parts)-1]
	if token == "" {
		token = "UNKNOWN"
	}
	ctx["policyNumber"] = "AUTO-" + token
	ctx["deductible"] = "$500"
	return map[string]any{"ok": true}, nil
}

func retrieveHomePolicyOptions(ctx map[string]any, args map[string]any) (map[string]any, error) {
	options := []session.Option{{Text: "Home-123"}, {Text: "Home-456"}}
	return map[string]any{"options": options}, nil
}

func validateHomeSelection(ctx map[string]any, args map[string]any) (map[string]any, error) {
	val, _ := args["value"].(string)
	if val == "" {
		return nil, errors.New("no home policy")
	}
	ctx["policyNumber"] = val
	ctx["deductible"] = "$1,000"
	return map[string]any{"ok": true}, nil
}

func getRenterPolicies(ctx map[string]any, args map[string]any) (map[string]any, error) {
	options := []session.Option{{Text: "Rent-111"}, {Text: "Rent-222"}}
	return map[string]any{"options": options}, nil
}

func validateRenterSelection(ctx map[string]any, args map[string]any) (map[string]any, error) {
	val, _ := args["value"].(string)
	if val == "" {
		return nil, errors.New("no renter policy")
	}
	ctx["policyNumber"] = val
	ctx["deductible"] = "$250"
	return map[string]any{"ok": true}, nil
}

func getPolicyPortalLink(ctx map[string]any, args map[string]any) (map[string]any, error) {
	num, _ := ctx["policyNumber"].(string)
	return map[string]any{"options": []map[string]any{{"text": "Open Portal", "url": "https://portal.example/policy/" + num}}}, nil
}

func getContactOptions(ctx map[string]any, args map[string]any) (map[string]any, error) {
	return map[string]any{"options": []map[string]any{{"text": "Email Support"}, {"text": "Call 1-800-555"}}}, nil
}

// Example of dynamic next step resolver used by autoPolicyDetail
func resolveNextStep(ctx map[string]any, args map[string]any) (map[string]any, error) {
	// In real logic, we might branch by vehicle coverage; here always go to deductibleResult
	return map[string]any{"nextStep": "deductibleResult"}, nil
}
