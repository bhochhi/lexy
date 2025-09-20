package nlu

import "strings"

// Client is the NLU service interface used by the orchestrator.
type Client interface {
	DetectIntent(text, clientID string) (intent string, slots map[string]string, err error)
}

// client is a simple in-memory NLU useful for local dev and tests.
type client struct{}

// NewClient constructs the default NLU client.
// Return the interface to hide the concrete type and avoid method-set issues.
func NewClient() Client { return &client{} }

func (c *client) DetectIntent(text, clientID string) (string, map[string]string, error) {
	low := strings.ToLower(text)
	slots := map[string]string{}
	switch {
	case strings.Contains(low, "dispute") || strings.Contains(low, "chargeback") || strings.Contains(low, "fraud"):
		return "disputeTransaction", map[string]string{}, nil
	case strings.Contains(low, "deductible"):
		if strings.Contains(low, "auto") {
			slots["policyType"] = "Auto Policy"
		} else if strings.Contains(low, "home") {
			slots["policyType"] = "Home Policy"
		} else if strings.Contains(low, "rent") {
			slots["policyType"] = "Renter Policy"
		}
		return "insurance-n-deductible-n-review", slots, nil
	default:
		return "fallback", nil, nil
	}
}
