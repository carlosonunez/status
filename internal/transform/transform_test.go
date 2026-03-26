package transform_test

import (
	"testing"
	"time"

	"github.com/carlosonunez/status/internal/getter"
	"github.com/carlosonunez/status/internal/transform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type applyTest struct {
	TestName   string
	Pattern    string
	Setters    map[string]transform.StatusTemplate
	Event      getter.Event
	SetterName string
	WantMatch  bool
	WantStatus transform.ResolvedStatus
}

func (tc *applyTest) Run(t *testing.T) {
	t.Helper()
	tr, err := transform.New(tc.Pattern, tc.Setters)
	require.NoError(t, err, "[%s] unexpected error building transform", tc.TestName)

	status, matched := tr.Apply(tc.Event, tc.SetterName)
	assert.Equal(t, tc.WantMatch, matched, "[%s] match mismatch", tc.TestName)
	if tc.WantMatch {
		assert.Equal(t, tc.WantStatus, status, "[%s] status mismatch", tc.TestName)
	}
}

func TestTransform_Apply(t *testing.T) {
	dur := 30 * time.Minute

	tests := []applyTest{
		{
			TestName: "simple_literal_match",
			Pattern:  "In a meeting",
			Setters: map[string]transform.StatusTemplate{
				"slack": {MessageTpl: "In a meeting", Emoji: "📆"},
			},
			Event:      getter.Event{Title: "In a meeting"},
			SetterName: "slack",
			WantMatch:  true,
			WantStatus: transform.ResolvedStatus{Message: "In a meeting", Emoji: "📆"},
		},
		{
			TestName: "wildcard_matches_any_title",
			Pattern:  ".*",
			Setters: map[string]transform.StatusTemplate{
				"slack": {MessageTpl: "Busy", Emoji: "🔴"},
			},
			Event:      getter.Event{Title: "Some random event"},
			SetterName: "slack",
			WantMatch:  true,
			WantStatus: transform.ResolvedStatus{Message: "Busy", Emoji: "🔴"},
		},
		{
			TestName: "pattern_no_match",
			Pattern:  "^FLIGHT",
			Setters: map[string]transform.StatusTemplate{
				"slack": {MessageTpl: "In flight", Emoji: "✈️"},
			},
			Event:      getter.Event{Title: "Team standup"},
			SetterName: "slack",
			WantMatch:  false,
		},
		{
			TestName: "capture_group_substitution",
			Pattern:  `^([A-Z]{2,3}[0-9]{1,4}) ([A-Z]{3}) to ([A-Z]{3})$`,
			Setters: map[string]transform.StatusTemplate{
				"slack": {MessageTpl: "In flight: $1: $2 -> $3", Emoji: "✈️"},
			},
			Event:      getter.Event{Title: "UAL1 JFK to LAX"},
			SetterName: "slack",
			WantMatch:  true,
			WantStatus: transform.ResolvedStatus{Message: "In flight: UAL1: JFK -> LAX", Emoji: "✈️"},
		},
		{
			TestName: "unknown_setter_name_no_match",
			Pattern:  ".*",
			Setters: map[string]transform.StatusTemplate{
				"slack": {MessageTpl: "Busy"},
			},
			Event:      getter.Event{Title: "Some event"},
			SetterName: "github",
			WantMatch:  false,
		},
		{
			TestName: "duration_and_ooo_preserved",
			Pattern:  "OOO.*",
			Setters: map[string]transform.StatusTemplate{
				"slack": {
					MessageTpl:    "On vacation",
					Emoji:         "🌴",
					Duration:      &dur,
					IsOutOfOffice: true,
				},
			},
			Event:      getter.Event{Title: "OOO - Beach trip"},
			SetterName: "slack",
			WantMatch:  true,
			WantStatus: transform.ResolvedStatus{
				Message:       "On vacation",
				Emoji:         "🌴",
				Duration:      &dur,
				IsOutOfOffice: true,
			},
		},
		{
			TestName: "dollar_n_adjacent_to_text_normalised",
			Pattern:  `^(.*) standup$`,
			Setters: map[string]transform.StatusTemplate{
				"slack": {MessageTpl: "In $1 standup", Emoji: "📆"},
			},
			Event:      getter.Event{Title: "Engineering standup"},
			SetterName: "slack",
			WantMatch:  true,
			WantStatus: transform.ResolvedStatus{Message: "In Engineering standup", Emoji: "📆"},
		},
		{
			TestName: "multiple_setters_correct_one_returned",
			Pattern:  ".*",
			Setters: map[string]transform.StatusTemplate{
				"slack":  {MessageTpl: "Slack status", Emoji: "💬"},
				"github": {MessageTpl: "GitHub status", Emoji: "🐙"},
			},
			Event:      getter.Event{Title: "Any event"},
			SetterName: "github",
			WantMatch:  true,
			WantStatus: transform.ResolvedStatus{Message: "GitHub status", Emoji: "🐙"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.TestName, tc.Run)
	}
}

func TestNew_InvalidRegex(t *testing.T) {
	_, err := transform.New(`[invalid`, map[string]transform.StatusTemplate{
		"slack": {MessageTpl: "test"},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid pattern")
}
