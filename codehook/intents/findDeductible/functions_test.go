package findDeductible_test

import (
	"testing"

	fd "github.com/bhochhi/lexy/codehook/intents/findDeductible"
)

// Happy path: valid policy selection updates context and returns ok
func TestValidatePolicySelection_SetsContextOnValid(t *testing.T) {
	hook := fd.New()
	ctx := map[string]any{}
	args := map[string]any{"value": "Home Policy"}

	res, err := hook.Invoke("validatePolicySelection", ctx, args)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if _, ok := res["ok"]; !ok {
		t.Fatalf("expected res['ok'] to be present, got %v", res)
	}
	if got := ctx["policyType"]; got != "Home Policy" {
		t.Fatalf("expected ctx['policyType'] to be 'Home Policy', got %v", got)
	}
}

// Error path: invalid selection returns error and does not update context
func TestValidatePolicySelection_Invalid(t *testing.T) {
	hook := fd.New()
	ctx := map[string]any{}
	args := map[string]any{"value": "NotARealPolicy"}

	if _, err := hook.Invoke("validatePolicySelection", ctx, args); err == nil {
		t.Fatalf("expected error for invalid selection, got nil")
	}
	if _, ok := ctx["policyType"]; ok {
		t.Fatalf("did not expect ctx['policyType'] to be set, got %v", ctx["policyType"])
	}
}
