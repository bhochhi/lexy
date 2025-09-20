package registry

// This file centralizes the blank imports of intent and domain packages to trigger their init() self-registrations.
// Developers adding a new intent/domain should add a new blank import here.

import (
	// Intents
	_ "github.com/bhochhi/lexy/codehook/intents/disputeTransaction"
	_ "github.com/bhochhi/lexy/codehook/intents/findDeductible"

	// Domains
	_ "github.com/bhochhi/lexy/codehook/domain/banking"
	_ "github.com/bhochhi/lexy/codehook/domain/insurance"
)
