package mocks

import (
	v1 "github.com/carlosonunez/status/api/v1"
)

// RenderedConfig is mocks/status.yaml, but rendered.
var RenderedConfig v1.Config = v1.Config{
	Sources: []v1.Source{
		{
			Name: "example",
			Settings: v1.SourceSettings{
				Enabled:      true,
				Weight:       0,
				PollDuration: "30m",
			},
			Properties: map[string]interface{}{
				"foo": "bar",
			},
			StatusGeneratingEvents: []v1.StatusGeneratingEvent{
				{
					Name:    "something that should update status",
					Enabled: true,
					Weight:  0,
					IncludeIf: []v1.EventIncludeRule{
						{
							RuleType: "regexp-matches",
							Rule:     ".*send a status now!$",
						},
					},
					AlwaysOverwrite: true,
					Transforms: []v1.EventTransform{
						{
							Input:    "send a (status) now!$",
							Template: "{{ .Source.Today }} is a {{ index .CaptureGroups 0 }}",
						},
					},
					LockDuration: v1.LockDuration{
						DefaultDuration: "30m",
					},
				},
			},
		},
	},
	Settings: v1.Settings{
		PubSub: v1.PubSub{
			Type: "example",
			Credentials: map[string]interface{}{
				"username": "foo",
				"password": "bar",
			},
			Properties: map[string]interface{}{
				"foo": "bar",
			},
		},
	},
}
