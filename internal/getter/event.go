package getter

import "time"

// Event represents a single calendar or external event retrieved by an EventGetter.
type Event struct {
	Calendar string
	Title    string
	StartsAt time.Time
	EndsAt   time.Time
	AllDay   bool
	// Meta holds arbitrary extra data an EventGetter may attach (e.g. flight number).
	Meta map[string]any
}
