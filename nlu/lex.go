package nlu

// For MVP we mock Amazon Lex: detect a single intent with a policyType slot if possible.

import "strings"

// DetectIntent returns mocked intent and slots based on simple keyword checks.
func DetectIntent(text, clientID string) (string, map[string]string, error) {
	intent := "findDeductible"
	slots := map[string]string{}
	low := strings.ToLower(text)
	if strings.Contains(low, "auto") {
		slots["policyType"] = "Auto Policy"
	} else if strings.Contains(low, "home") {
		slots["policyType"] = "Home Policy"
	} else if strings.Contains(low, "rent") {
		slots["policyType"] = "Renter Policy"
	}
	return intent, slots, nil
}
