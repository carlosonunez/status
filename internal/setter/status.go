package setter

import "time"

// Status is the resolved status message to be sent to a StatusSetter.
type Status struct {
	Message       string
	Emoji         string
	Duration      *time.Duration
	IsOutOfOffice bool
}

// SetResult is returned by a StatusSetter after attempting to apply a Status.
type SetResult struct {
	Skipped  bool
	Response string
}
