package setter

// Status is the resolved set of parameters to be sent to a StatusSetter.
// Each StatusSetter implementation extracts the fields it supports from Params.
type Status struct {
	// Params holds setter-specific key/value pairs (e.g. "status_message", "emoji").
	// String values may have already had $N capture-group references substituted
	// by the transform layer.
	Params map[string]any
}

// SetResult is returned by a StatusSetter after attempting to apply a Status.
type SetResult struct {
	Skipped  bool
	Response string
}
