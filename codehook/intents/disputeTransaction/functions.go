package disputeTransaction

import (
	"errors"
	"fmt"
	"strings"

	"github.com/bhochhi/lexy/session"
)

type Registry struct{}

func NewRegistry() *Registry { return &Registry{} }

func (r *Registry) Call(name string, ctx map[string]any, args map[string]any) (map[string]any, error) {
	switch name {
	case "checkClientAccounts":
		return checkClientAccounts(ctx, args)
	case "getAccounts":
		return getAccounts(ctx, args)
	case "validateAccountSelection":
		return validateAccountSelection(ctx, args)
	case "resolveDisputePath":
		return resolveDisputePath(ctx, args)
	case "getDelinquencyPortalLink":
		return getDelinquencyPortalLink(ctx, args)
	case "getDisputeFormLink":
		return getDisputeFormLink(ctx, args)
	case "getContactOptions":
		return getContactOptions(ctx, args)
	default:
		return nil, fmt.Errorf("unknown function: %s", name)
	}
}

// Registration is manual via registry.New(); no init-based registration here.

func checkClientAccounts(ctx map[string]any, args map[string]any) (map[string]any, error) {
	// Mock: if clientId ends with 9, no accounts
	cid, _ := ctx["clientId"].(string)
	if strings.HasSuffix(cid, "9") {
		return map[string]any{"hasData": false, "hasNoData": true}, nil
	}
	return map[string]any{"hasData": true, "hasNoData": false}, nil
}

func getAccounts(ctx map[string]any, args map[string]any) (map[string]any, error) {
	// Mock accounts list
	options := []session.Option{
		{Text: "Credit Card •••• 4242"},
		{Text: "Debit Card •••• 1111"},
		{Text: "Checking •••• 2001"},
		{Text: "Savings •••• 3001"},
	}
	return map[string]any{"options": options}, nil
}

func validateAccountSelection(ctx map[string]any, args map[string]any) (map[string]any, error) {
	val, _ := args["value"].(string)
	if val == "" {
		return nil, errors.New("no account")
	}
	ctx["accountId"] = lastToken(val)
	if strings.Contains(strings.ToLower(val), "credit") {
		ctx["accountType"] = "Credit Card"
	} else if strings.Contains(strings.ToLower(val), "debit") {
		ctx["accountType"] = "Debit Card"
	} else if strings.Contains(strings.ToLower(val), "checking") {
		ctx["accountType"] = "Checking"
	} else if strings.Contains(strings.ToLower(val), "savings") {
		ctx["accountType"] = "Savings"
	} else {
		ctx["accountType"] = "Account"
	}
	return map[string]any{"ok": true}, nil
}

func resolveDisputePath(ctx map[string]any, args map[string]any) (map[string]any, error) {
	// Mock: 20% of credit cards are delinquent
	accType, _ := ctx["accountType"].(string)
	accId, _ := ctx["accountId"].(string)
	if accType == "Credit Card" && strings.HasSuffix(accId, "1") {
		return map[string]any{"nextStep": "delinquencyNotice"}, nil
	}
	return map[string]any{"nextStep": "disputeForm"}, nil
}

func getDelinquencyPortalLink(ctx map[string]any, args map[string]any) (map[string]any, error) {
	id, _ := ctx["accountId"].(string)
	return map[string]any{"options": []map[string]any{{"text": "Resolve Delinquency", "url": "https://bank.example/accounts/" + id + "/delinquency"}}}, nil
}

func getDisputeFormLink(ctx map[string]any, args map[string]any) (map[string]any, error) {
	id, _ := ctx["accountId"].(string)
	return map[string]any{"options": []map[string]any{{"text": "Open Dispute Form", "url": "https://bank.example/accounts/" + id + "/dispute"}}}, nil
}

func getContactOptions(ctx map[string]any, args map[string]any) (map[string]any, error) {
	return map[string]any{"options": []map[string]any{{"text": "Email Support"}, {"text": "Call 1-800-555"}}}, nil
}

func lastToken(s string) string {
	parts := strings.Fields(s)
	if len(parts) == 0 {
		return "UNK"
	}
	return parts[len(parts)-1]
}
