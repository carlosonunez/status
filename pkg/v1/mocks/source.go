package mocks

import v1 "github.com/carlosonunez/status/api/v1"

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
		},
		{
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
