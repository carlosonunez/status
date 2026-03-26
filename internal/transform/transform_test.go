package transform_test

import (
	"testing"

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
	WantParams map[string]any
}

func (tc *applyTest) Run(t *testing.T) {
	t.Helper()
	tr, err := transform.New(tc.Pattern, tc.Setters)
	require.NoError(t, err, "[%s] unexpected error building transform", tc.TestName)

	status, matched := tr.Apply(tc.Event, tc.SetterName)
	assert.Equal(t, tc.WantMatch, matched, "[%s] match mismatch", tc.TestName)
	if tc.WantMatch {
		assert.Equal(t, tc.WantParams, status.Params, "[%s] params mismatch", tc.TestName)
	}
}

func TestTransform_Apply(t *testing.T) {
	tests := []applyTest{
		{
			TestName: "simple_literal_match",
			Pattern:  "In a meeting",
			Setters: map[string]transform.StatusTemplate{
				"slack": {Params: map[string]any{"status_message": "In a meeting", "emoji": "📆"}},
			},
			Event:      getter.Event{Title: "In a meeting"},
			SetterName: "slack",
			WantMatch:  true,
			WantParams: map[string]any{"status_message": "In a meeting", "emoji": "📆"},
		},
		{
			TestName: "wildcard_matches_any_title",
			Pattern:  ".*",
			Setters: map[string]transform.StatusTemplate{
				"slack": {Params: map[string]any{"status_message": "Busy", "emoji": "🔴"}},
			},
			Event:      getter.Event{Title: "Some random event"},
			SetterName: "slack",
			WantMatch:  true,
			WantParams: map[string]any{"status_message": "Busy", "emoji": "🔴"},
		},
		{
			TestName: "pattern_no_match",
			Pattern:  "^FLIGHT",
			Setters: map[string]transform.StatusTemplate{
				"slack": {Params: map[string]any{"status_message": "In flight"}},
			},
			Event:      getter.Event{Title: "Team standup"},
			SetterName: "slack",
			WantMatch:  false,
		},
		{
			TestName: "capture_group_substitution",
			Pattern:  `^([A-Z]{2,3}[0-9]{1,4}) ([A-Z]{3}) to ([A-Z]{3})$`,
			Setters: map[string]transform.StatusTemplate{
				"slack": {Params: map[string]any{
					"status_message": "In flight: $1: $2 -> $3",
					"emoji":          "✈️",
				}},
			},
			Event:      getter.Event{Title: "UAL1 JFK to LAX"},
			SetterName: "slack",
			WantMatch:  true,
			WantParams: map[string]any{
				"status_message": "In flight: UAL1: JFK -> LAX",
				"emoji":          "✈️",
			},
		},
		{
			TestName: "unknown_setter_name_no_match",
			Pattern:  ".*",
			Setters: map[string]transform.StatusTemplate{
				"slack": {Params: map[string]any{"status_message": "Busy"}},
			},
			Event:      getter.Event{Title: "Some event"},
			SetterName: "github",
			WantMatch:  false,
		},
		{
			TestName: "non_string_params_passed_through_unchanged",
			Pattern:  "OOO.*",
			Setters: map[string]transform.StatusTemplate{
				"slack": {Params: map[string]any{
					"status_message":   "On vacation",
					"emoji":            "🌴",
					"is_out_of_office": true,
					"duration":         "8h",
				}},
			},
			Event:      getter.Event{Title: "OOO - Beach trip"},
			SetterName: "slack",
			WantMatch:  true,
			WantParams: map[string]any{
				"status_message":   "On vacation",
				"emoji":            "🌴",
				"is_out_of_office": true,
				"duration":         "8h",
			},
		},
		{
			TestName: "dollar_n_adjacent_to_text_normalised",
			Pattern:  `^(.*) standup$`,
			Setters: map[string]transform.StatusTemplate{
				"slack": {Params: map[string]any{"status_message": "In $1 standup"}},
			},
			Event:      getter.Event{Title: "Engineering standup"},
			SetterName: "slack",
			WantMatch:  true,
			WantParams: map[string]any{"status_message": "In Engineering standup"},
		},
		{
			TestName: "multiple_setters_correct_one_returned",
			Pattern:  ".*",
			Setters: map[string]transform.StatusTemplate{
				"slack":  {Params: map[string]any{"status_message": "Slack status", "emoji": "💬"}},
				"github": {Params: map[string]any{"message": "GitHub status", "emoji": "🐙"}},
			},
			Event:      getter.Event{Title: "Any event"},
			SetterName: "github",
			WantMatch:  true,
			WantParams: map[string]any{"message": "GitHub status", "emoji": "🐙"},
		},
		{
			TestName: "setter_with_no_params_returns_empty_map",
			Pattern:  ".*",
			Setters: map[string]transform.StatusTemplate{
				"dummy": {Params: map[string]any{}},
			},
			Event:      getter.Event{Title: "Any event"},
			SetterName: "dummy",
			WantMatch:  true,
			WantParams: map[string]any{},
		},
		{
			TestName: "substitution_applied_to_all_string_params",
			Pattern:  `^(.+): (.+)$`,
			Setters: map[string]transform.StatusTemplate{
				"custom": {Params: map[string]any{
					"title":    "$1",
					"subtitle": "$2",
					"count":    42,
				}},
			},
			Event:      getter.Event{Title: "Project: Alpha"},
			SetterName: "custom",
			WantMatch:  true,
			WantParams: map[string]any{
				"title":    "Project",
				"subtitle": "Alpha",
				"count":    42,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.TestName, tc.Run)
	}
}

func TestNew_InvalidRegex(t *testing.T) {
	_, err := transform.New(`[invalid`, map[string]transform.StatusTemplate{
		"slack": {Params: map[string]any{"status_message": "test"}},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid pattern")
}
