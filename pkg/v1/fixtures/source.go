package fixtures

import (
	"time"

	v1 "github.com/carlosonunez/status/api/v1"
)

// ExampleSource is an example of a source with events.
type ExampleSource struct{}

// Initialize initializes the ExampleSource.
func (s *ExampleSource) Initialize(opts map[string]interface{}) error {
	return nil
}

// GetParent gets its config.
func (s *ExampleSource) GetParent() *v1.Source {
	return &exampleSourceConfig
}

// Poll discovers and returns new events.
func (s *ExampleSource) Poll(start *time.Time, end *time.Time) (*[]*v1.Event, error) {
	var evts []*v1.Event
	evtsProduced := []v1.Event{
		{
			Message: "SELECT_ME: But I don't need a transform.",
		},
		{
			Message:      "SELECT_ME_TOO: But I also don't need a transform.",
			OverrideLock: true,
		},
		{
			Message:      "Don't select me.",
			OverrideLock: true,
		},
		{
			Message: "SELECT_ME: Event: My event happened.",
		},
	}
	for _, evt := range evtsProduced {
		evts = append(evts, &evt)
	}
	return &evts, nil
}

// Today returns today's date (mocked)
func (s *ExampleSource) Today() string {
	return "Mar 4, 2022"
}

var exampleSourceConfig v1.Source = v1.Source{
	Name: "example",
	Settings: v1.SourceSettings{
		Enabled:      true,
		Weight:       15,
		PollDuration: "1m",
	},
	StatusGeneratingEvents: []v1.StatusGeneratingEvent{
		{
			Name:    "first-event",
			Enabled: true,
			Weight:  40,
			IncludeIf: []v1.EventIncludeRule{
				{
					RuleType: "matches-regexp",
					Rule:     "^SELECT_ME_TOO:",
				},
			},
			Transforms: []v1.EventTransform{
				{
					Input:    ".* Event: (My event).*$",
					Template: "{{ index .CaptureGroups 0 }} happened on {{ .Source.Today }}",
				},
			},
		},
		{
			Name:    "second-event",
			Enabled: true,
			Weight:  32,
			IncludeIf: []v1.EventIncludeRule{
				{
					RuleType: "matches-regexp",
					Rule:     "^SELECT_ME: .*Event: My event",
				},
			},
			Transforms: []v1.EventTransform{
				{
					Input:    ".* Event: (My event).*$",
					Template: "{{ index .CaptureGroups 0 }} happened on {{ .Source.Today }}",
				},
			},
		},
	},
}
