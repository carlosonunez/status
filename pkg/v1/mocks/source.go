package mocks

import (
	"fmt"
	"math/rand"
	"time"

	v1 "github.com/carlosonunez/status/api/v1"
)

// SourceWithInvalidEventRules is an example of a source with an invalid rule
// set.
var SourceWithInvalidEventRules v1.Source = v1.Source{
	Name: "invalid-source",
	StatusGeneratingEvents: []v1.StatusGeneratingEvent{
		{
			Name: "valid-event-1",
			IncludeIf: []v1.EventIncludeRule{
				{
					RuleType: "valid-rule",
					Rule:     "foo",
				},
				{
					RuleType: "valid-rule-2",
					Rule:     "foo",
				},
			},
		}, {
			Name: "valid-event-2",
			IncludeIf: []v1.EventIncludeRule{
				{
					RuleType: "valid-rule",
					Rule:     "foo",
				},
				{
					RuleType: "invalid-rule",
					Rule:     "foo",
				},
			},
		},
	},
}

// MockSource is used to test against the interface.
type MockSource struct{}

// Initialize inits the MockSource
func (c *MockSource) Initialize(props map[string]interface{}) error {
	return nil
}

// Poll simulates an event poll.
func (c *MockSource) Poll() ([]*v1.Event, error) {
	rand.Seed(time.Now().UnixNano())
	r := rand.Intn(500)
	// Simulate a small amount of network latency.
	time.Sleep(time.Duration(r) * time.Millisecond)
	numEvents := rand.Intn(10)
	var evts []*v1.Event
	for idx := 0; idx < numEvents; idx++ {
		evt := v1.Event{
			Message: fmt.Sprintf("this is event %d", idx),
		}
		evts = append(evts, &evt)
	}
	return evts, nil
}

// ParseDate takes a date inside of an event message and returns a
// fake date.
func (c *MockSource) ParseDate(s string) string {
	return "Mar 4, 2022"
}

// ReallyLongTransformFunction is a function that takes a really long time.
// It should be cut off after TransformEventTimeoutSeconds.
func (c *MockSource) ReallyLongTransformFunction(s string) string {
	time.Sleep(10000000000000)
	return "shouldn't see this"
}
