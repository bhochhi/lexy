package codehook

// Hook is the contract for invoking named codehook functions with context and args.
// Implementations typically dispatch by function name to internal handlers and
// return results as a generic map. Errors indicate failure paths for the orchestrator.
type Hook interface {
	Invoke(name string, ctx map[string]any, args map[string]any) (map[string]any, error)
}
