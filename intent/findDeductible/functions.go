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
	switch name {
	case "checkClientPolicies":
		return fnCheckClientPolicies(ctx, args)
	case "getDynamicPolicies":
		return fnGetDynamicPolicies(ctx, args)
	case "validatePolicySelection":
		return fnValidatePolicySelection(ctx, args)
	case "getClientVehicles":
		return fnGetClientVehicles(ctx, args)
	case "validateVehicleSelection":
		return fnValidateVehicleSelection(ctx, args)
	case "getHomePolicies":
		return fnGetHomePolicies(ctx, args)
	case "validateHomeSelection":
		return fnValidateHomeSelection(ctx, args)
	case "getRenterPolicies":
		return fnGetRenterPolicies(ctx, args)
	case "validateRenterSelection":
		return fnValidateRenterSelection(ctx, args)
	case "getPolicyPortalLink":
		return fnGetPolicyPortalLink(ctx, args)
	case "getContactOptions":
		return fnGetContactOptions(ctx, args)
	case "resolveNextStepFunction":
		return fnResolveNextStep(ctx, args)
	default:
		return nil, fmt.Errorf("unknown function: %s", name)
	}
}

// --- Function implementations (mocked) ---

func fnCheckClientPolicies(ctx map[string]any, args map[string]any) (map[string]any, error) {
	// Pretend to check via API; here we mock that client has data unless clientId ends with 0
	cid, _ := ctx["clientId"].(string)
	if strings.HasSuffix(cid, "0") {
		return map[string]any{"hasData": false, "hasNoData": true}, nil
	}
	return map[string]any{"hasData": true, "hasNoData": false}, nil
}

func fnGetDynamicPolicies(ctx map[string]any, args map[string]any) (map[string]any, error) {
	// Build options dynamically; in real code, query user policies
	options := []session.Option{
		{Text: "Umbrella Policy"},
	}
	return map[string]any{"options": options}, nil
}

func fnValidatePolicySelection(ctx map[string]any, args map[string]any) (map[string]any, error) {
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

func fnGetClientVehicles(ctx map[string]any, args map[string]any) (map[string]any, error) {
	// Mock two vehicles
	options := []session.Option{
		{Text: "2019 Honda Civic - ABC123"},
		{Text: "2021 Tesla Model 3 - XYZ789"},
	}
	return map[string]any{"options": options}, nil
}

func fnValidateVehicleSelection(ctx map[string]any, args map[string]any) (map[string]any, error) {
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

func fnGetHomePolicies(ctx map[string]any, args map[string]any) (map[string]any, error) {
	options := []session.Option{{Text: "Home-123"}, {Text: "Home-456"}}
	return map[string]any{"options": options}, nil
}

func fnValidateHomeSelection(ctx map[string]any, args map[string]any) (map[string]any, error) {
	val, _ := args["value"].(string)
	if val == "" {
		return nil, errors.New("no home policy")
	}
	ctx["policyNumber"] = val
	ctx["deductible"] = "$1,000"
	return map[string]any{"ok": true}, nil
}

func fnGetRenterPolicies(ctx map[string]any, args map[string]any) (map[string]any, error) {
	options := []session.Option{{Text: "Rent-111"}, {Text: "Rent-222"}}
	return map[string]any{"options": options}, nil
}

func fnValidateRenterSelection(ctx map[string]any, args map[string]any) (map[string]any, error) {
	val, _ := args["value"].(string)
	if val == "" {
		return nil, errors.New("no renter policy")
	}
	ctx["policyNumber"] = val
	ctx["deductible"] = "$250"
	return map[string]any{"ok": true}, nil
}

func fnGetPolicyPortalLink(ctx map[string]any, args map[string]any) (map[string]any, error) {
	num, _ := ctx["policyNumber"].(string)
	return map[string]any{"options": []map[string]any{{"text": "Open Portal", "url": "https://portal.example/policy/" + num}}}, nil
}

func fnGetContactOptions(ctx map[string]any, args map[string]any) (map[string]any, error) {
	return map[string]any{"options": []map[string]any{{"text": "Email Support"}, {"text": "Call 1-800-555"}}}, nil
}

// Example of dynamic next step resolver used by autoPolicyDetail
func fnResolveNextStep(ctx map[string]any, args map[string]any) (map[string]any, error) {
	// In real logic, we might branch by vehicle coverage; here always go to deductibleResult
	return map[string]any{"nextStep": "deductibleResult"}, nil
}
