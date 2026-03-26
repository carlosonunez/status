package setter

import "github.com/carlosonunez/status/internal/params"

// Status is the resolved set of parameters to be sent to a StatusSetter.
// Each StatusSetter implementation uses Params accessors to extract the
// fields it supports.
type Status struct {
	Params params.Params
}

// SetResult is returned by a StatusSetter after attempting to apply a Status.
type SetResult struct {
	Skipped  bool
	Response string
}
