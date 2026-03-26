package rule_test

import (
	"testing"

	"github.com/carlosonunez/status/internal/config"
	"github.com/carlosonunez/status/internal/rule"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fromConfigTest struct {
	TestName    string
	Input       config.RuleConfig
	WantErr     bool
	WantErrText string
	WantName    string
	WantGetter  string
	WantTransformCount int
}

func (tc *fromConfigTest) Run(t *testing.T) {
	t.Helper()
	r, err := rule.FromConfig(tc.Input)
	if tc.WantErr {
		require.Error(t, err, "[%s] expected error", tc.TestName)
		assert.Contains(t, err.Error(), tc.WantErrText, "[%s] error text mismatch", tc.TestName)
		return
	}
	require.NoError(t, err, "[%s] unexpected error", tc.TestName)
	assert.Equal(t, tc.WantName, r.Name, "[%s] name mismatch", tc.TestName)
	assert.Equal(t, tc.WantGetter, r.GetterName, "[%s] getter name mismatch", tc.TestName)
	assert.Len(t, r.Transforms, tc.WantTransformCount, "[%s] transform count mismatch", tc.TestName)
}

func TestFromConfig(t *testing.T) {
	tests := []fromConfigTest{
		{
			TestName: "minimal_valid_rule",
			Input: config.RuleConfig{
				Name:        "My rule",
				EventGetter: config.EventGetterConfig{Name: "google_calendar"},
				Transforms: []config.TransformConfig{
					{
						Name:    "catch-all",
						Pattern: ".*",
						StatusSetters: []config.StatusSetterConfig{
							{Name: "slack", Params: map[string]interface{}{
								"status_message": "In a meeting",
								"emoji":          "📆",
							}},
						},
					},
				},
			},
			WantName:           "My rule",
			WantGetter:         "google_calendar",
			WantTransformCount: 1,
		},
		{
			TestName: "multiple_transforms_preserved_in_order",
			Input: config.RuleConfig{
				Name:        "Ordered rule",
				EventGetter: config.EventGetterConfig{Name: "google_calendar"},
				Transforms: []config.TransformConfig{
					{Name: "first", Pattern: "^FLIGHT", StatusSetters: []config.StatusSetterConfig{{Name: "slack"}}},
					{Name: "second", Pattern: ".*", StatusSetters: []config.StatusSetterConfig{{Name: "slack"}}},
				},
			},
			WantName:           "Ordered rule",
			WantGetter:         "google_calendar",
			WantTransformCount: 2,
		},
		{
			TestName: "invalid_regex_returns_error",
			Input: config.RuleConfig{
				Name:        "Bad rule",
				EventGetter: config.EventGetterConfig{Name: "google_calendar"},
				Transforms: []config.TransformConfig{
					{Name: "bad", Pattern: "[invalid", StatusSetters: []config.StatusSetterConfig{{Name: "slack"}}},
				},
			},
			WantErr:     true,
			WantErrText: "invalid pattern",
		},
		{
			TestName: "getter_params_preserved",
			Input: config.RuleConfig{
				Name: "Rule with params",
				EventGetter: config.EventGetterConfig{
					Name:   "google_calendar",
					Params: map[string]interface{}{"allow_all_day_events": true},
				},
				Transforms: []config.TransformConfig{
					{Name: "t", Pattern: ".*", StatusSetters: []config.StatusSetterConfig{{Name: "slack"}}},
				},
			},
			WantName:           "Rule with params",
			WantGetter:         "google_calendar",
			WantTransformCount: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.TestName, tc.Run)
	}
}

func TestSortByPriority(t *testing.T) {
	rules := []*rule.Rule{
		{Name: "third", Priority: 3},
		{Name: "first", Priority: 1},
		{Name: "second", Priority: 2},
	}
	rule.SortByPriority(rules)

	assert.Equal(t, "first", rules[0].Name)
	assert.Equal(t, "second", rules[1].Name)
	assert.Equal(t, "third", rules[2].Name)
}

func TestSortByPriority_ZeroPriorityLast(t *testing.T) {
	rules := []*rule.Rule{
		{Name: "no-priority", Priority: 0},
		{Name: "high", Priority: 1},
	}
	rule.SortByPriority(rules)

	assert.Equal(t, "high", rules[0].Name)
	assert.Equal(t, "no-priority", rules[1].Name)
}

func TestTransformName_PreservedFromConfig(t *testing.T) {
	cfg := config.RuleConfig{
		Name:        "My rule",
		EventGetter: config.EventGetterConfig{Name: "google_calendar"},
		Transforms: []config.TransformConfig{
			{Name: "In a flight", Pattern: `^([A-Z]+\d+)`, StatusSetters: []config.StatusSetterConfig{{Name: "slack"}}},
		},
	}
	r, err := rule.FromConfig(cfg)
	require.NoError(t, err)
	assert.Equal(t, "In a flight", r.Transforms[0].Name)
}
