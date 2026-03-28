package engine_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/carlosonunez/status/internal/config"
	"github.com/carlosonunez/status/internal/engine"
	"github.com/carlosonunez/status/internal/getter"
	"github.com/carlosonunez/status/internal/params"
	"github.com/carlosonunez/status/internal/pluginspec"
	"github.com/carlosonunez/status/internal/rule"
	"github.com/carlosonunez/status/internal/setter"
	"github.com/carlosonunez/status/internal/transform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- fakes ---

type fakeGetter struct {
	name   string
	events []getter.Event
	err    error
}

func (f *fakeGetter) Name() string                          { return f.name }
func (f *fakeGetter) MinInterval() time.Duration            { return time.Second }
func (f *fakeGetter) ParamSpecs() []pluginspec.ParamSpec    { return nil }
func (f *fakeGetter) GetEvents(_ context.Context, _ params.Params) ([]getter.Event, error) {
	return f.events, f.err
}

type fakeSetter struct {
	mu       sync.Mutex
	name     string
	received []setter.Status
}

func (f *fakeSetter) Name() string                       { return f.name }
func (f *fakeSetter) ParamSpecs() []pluginspec.ParamSpec { return nil }
func (f *fakeSetter) SetStatus(_ context.Context, s setter.Status) (setter.SetResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.received = append(f.received, s)
	return setter.SetResult{Response: "ok"}, nil
}
func (f *fakeSetter) Received() []setter.Status {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]setter.Status, len(f.received))
	copy(out, f.received)
	return out
}

// --- helpers ---

// fixedNow is the reference time used across tests.
var fixedNow = time.Date(2026, 3, 26, 14, 0, 0, 0, time.UTC)

// activeEvent returns an event that is active at fixedNow.
func activeEvent(title string) getter.Event {
	return getter.Event{
		Title:    title,
		StartsAt: fixedNow.Add(-time.Hour),
		EndsAt:   fixedNow.Add(time.Hour),
	}
}

// futureEvent returns an event that has not started yet at fixedNow.
func futureEvent(title string) getter.Event {
	return getter.Event{
		Title:    title,
		StartsAt: fixedNow.Add(time.Hour),
		EndsAt:   fixedNow.Add(2 * time.Hour),
	}
}

// pastEvent returns an event that has already ended at fixedNow.
func pastEvent(title string) getter.Event {
	return getter.Event{
		Title:    title,
		StartsAt: fixedNow.Add(-2 * time.Hour),
		EndsAt:   fixedNow.Add(-time.Hour),
	}
}

// makeRule builds a Rule with a single transform pattern → one setter.
func makeRule(name string, priority int, getterName, pattern, setterName, msg string) *rule.Rule {
	setters := map[string]transform.StatusTemplate{
		setterName: {Params: params.FromMap(map[string]any{"status_message": msg})},
	}
	tr, _ := transform.New(pattern, setters)
	tr.Name = "t"
	return &rule.Rule{
		Name:       name,
		Priority:   priority,
		GetterName: getterName,
		Transforms: []*transform.Transform{tr},
	}
}

// makeRuleMultiTransform builds a Rule with two ordered transforms.
func makeRuleMultiTransform(name string, getterName, setterName string, transforms ...[2]string) *rule.Rule {
	trs := make([]*transform.Transform, 0, len(transforms))
	for _, pair := range transforms {
		pattern, msg := pair[0], pair[1]
		setterMap := map[string]transform.StatusTemplate{
			setterName: {Params: params.FromMap(map[string]any{"status_message": msg})},
		}
		tr, _ := transform.New(pattern, setterMap)
		trs = append(trs, tr)
	}
	return &rule.Rule{
		Name:       name,
		GetterName: getterName,
		Transforms: trs,
	}
}

// --- processGroup tests ---

type processGroupTest struct {
	TestName   string
	Events     []getter.Event
	GetterErr  error
	Rules      []*rule.Rule
	SetterName string
	WantCount  int
	CheckFirst func(t *testing.T, s setter.Status)
}

func (tc processGroupTest) RunTest(t *testing.T) {
	t.Helper()
	fs := &fakeSetter{name: tc.SetterName}
	sr := setter.NewRegistry()
	sr.Register(fs)

	fg := &fakeGetter{name: "fake", events: tc.Events, err: tc.GetterErr}

	e := engine.NewForTest(sr, func() time.Time { return fixedNow })
	e.ProcessGroup(context.Background(), fg, params.FromMap(nil), tc.Rules)

	received := fs.Received()
	assert.Len(t, received, tc.WantCount, "wrong number of setter calls")
	if tc.CheckFirst != nil && len(received) > 0 {
		tc.CheckFirst(t, received[0])
	}
}

func TestEngine_ProcessGroup(t *testing.T) {
	tests := []processGroupTest{
		{
			TestName:   "active_event_dispatches_status",
			Events:     []getter.Event{activeEvent("Team standup")},
			Rules:      []*rule.Rule{makeRule("r", 1, "fake", ".*", "fake", "busy")},
			SetterName: "fake",
			WantCount:  1,
			CheckFirst: func(t *testing.T, s setter.Status) {
				msg, ok := s.Params.String("status_message")
				require.True(t, ok)
				assert.Equal(t, "busy", msg)
			},
		},
		{
			TestName:   "future_event_not_dispatched",
			Events:     []getter.Event{futureEvent("Future meeting")},
			Rules:      []*rule.Rule{makeRule("r", 1, "fake", ".*", "fake", "busy")},
			SetterName: "fake",
			WantCount:  0,
		},
		{
			TestName:   "past_event_not_dispatched",
			Events:     []getter.Event{pastEvent("Past meeting")},
			Rules:      []*rule.Rule{makeRule("r", 1, "fake", ".*", "fake", "busy")},
			SetterName: "fake",
			WantCount:  0,
		},
		{
			TestName: "allday_event_active_today",
			Events: []getter.Event{{
				Title:    "All day event",
				StartsAt: time.Date(2026, 3, 26, 0, 0, 0, 0, time.UTC),
				AllDay:   true,
			}},
			Rules:      []*rule.Rule{makeRule("r", 1, "fake", ".*", "fake", "ooo")},
			SetterName: "fake",
			WantCount:  1,
		},
		{
			TestName: "allday_event_wrong_day_not_dispatched",
			Events: []getter.Event{{
				Title:    "Yesterday's event",
				StartsAt: time.Date(2026, 3, 25, 0, 0, 0, 0, time.UTC),
				AllDay:   true,
			}},
			Rules:      []*rule.Rule{makeRule("r", 1, "fake", ".*", "fake", "ooo")},
			SetterName: "fake",
			WantCount:  0,
		},
		{
			TestName:   "getter_error_no_dispatch",
			Events:     nil,
			GetterErr:  errors.New("calendar unavailable"),
			Rules:      []*rule.Rule{makeRule("r", 1, "fake", ".*", "fake", "busy")},
			SetterName: "fake",
			WantCount:  0,
		},
		{
			TestName: "priority_ordering_higher_priority_wins",
			Events:   []getter.Event{activeEvent("Team standup")},
			Rules: []*rule.Rule{
				makeRule("high", 1, "fake", ".*", "fake", "high priority"),
				makeRule("low", 2, "fake", ".*", "fake", "low priority"),
			},
			SetterName: "fake",
			WantCount:  1,
			CheckFirst: func(t *testing.T, s setter.Status) {
				msg, _ := s.Params.String("status_message")
				assert.Equal(t, "high priority", msg)
			},
		},
		{
			// Both transforms match the event; only the first should fire.
			TestName: "first_transform_wins_per_setter",
			Events:   []getter.Event{activeEvent("Team standup")},
			Rules: []*rule.Rule{
				makeRuleMultiTransform("r", "fake", "fake",
					[2]string{"^Team standup$", "In a meeting"},
					[2]string{".*", "Busy"},
				),
			},
			SetterName: "fake",
			WantCount:  1,
			CheckFirst: func(t *testing.T, s setter.Status) {
				msg, _ := s.Params.String("status_message")
				assert.Equal(t, "In a meeting", msg)
			},
		},
		{
			TestName: "unmatched_pattern_no_dispatch",
			Events:   []getter.Event{activeEvent("Team standup")},
			Rules:    []*rule.Rule{makeRule("r", 1, "fake", "^FLIGHT", "fake", "in flight")},
			SetterName: "fake",
			WantCount:  0,
		},
	}
	for _, tc := range tests {
		t.Run(tc.TestName, func(t *testing.T) {
			tc.RunTest(t)
		})
	}
}

// --- New validation tests ---

type newEngineTest struct {
	TestName    string
	Config      *config.Config
	SetupRegs   func(gr *getter.Registry, sr *setter.Registry)
	WantErr     bool
	ErrContains string
}

func (tc newEngineTest) RunTest(t *testing.T) {
	t.Helper()
	gr := getter.NewRegistry()
	sr := setter.NewRegistry()
	if tc.SetupRegs != nil {
		tc.SetupRegs(gr, sr)
	}
	eng, err := engine.New(tc.Config, gr, sr)
	if tc.WantErr {
		require.Error(t, err, "expected an error")
		if tc.ErrContains != "" {
			assert.Contains(t, err.Error(), tc.ErrContains)
		}
		assert.Nil(t, eng)
	} else {
		require.NoError(t, err)
		assert.NotNil(t, eng)
	}
}

func cfgWithRule(getterName, setterName string) *config.Config {
	return &config.Config{
		Rules: []config.RuleConfig{{
			Name:     "r",
			Priority: 1,
			EventGetter: config.EventGetterConfig{Name: getterName},
			Transforms: []config.TransformConfig{{
				Name:    "t",
				Pattern: ".*",
				StatusSetters: []config.StatusSetterConfig{{
					Name:   setterName,
					Params: map[string]any{"status_message": "busy"},
				}},
			}},
		}},
	}
}

func TestEngine_New(t *testing.T) {
	tests := []newEngineTest{
		{
			TestName: "valid_config_returns_engine",
			Config:   cfgWithRule("fake-getter", "fake-setter"),
			SetupRegs: func(gr *getter.Registry, sr *setter.Registry) {
				gr.Register(&fakeGetter{name: "fake-getter"})
				sr.Register(&fakeSetter{name: "fake-setter"})
			},
			WantErr: false,
		},
		{
			TestName:    "missing_getter_returns_error",
			Config:      cfgWithRule("missing-getter", "fake-setter"),
			SetupRegs:   func(gr *getter.Registry, sr *setter.Registry) {
				sr.Register(&fakeSetter{name: "fake-setter"})
			},
			WantErr:     true,
			ErrContains: "missing-getter",
		},
		{
			TestName: "missing_setter_returns_error",
			Config:   cfgWithRule("fake-getter", "missing-setter"),
			SetupRegs: func(gr *getter.Registry, sr *setter.Registry) {
				gr.Register(&fakeGetter{name: "fake-getter"})
			},
			WantErr:     true,
			ErrContains: "missing-setter",
		},
		{
			TestName:  "empty_config_returns_engine",
			Config:    &config.Config{},
			WantErr:   false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.TestName, func(t *testing.T) {
			tc.RunTest(t)
		})
	}
}

// --- Run / cancel test ---

func TestEngine_Run_StopsOnContextCancel(t *testing.T) {
	gr := getter.NewRegistry()
	gr.Register(&fakeGetter{name: "fake-getter"})
	sr := setter.NewRegistry()
	sr.Register(&fakeSetter{name: "fake-setter"})

	eng, err := engine.New(cfgWithRule("fake-getter", "fake-setter"), gr, sr)
	require.NoError(t, err)

	// Use a ticker that never fires so processGroup is never called.
	neverTick := make(chan time.Time)
	engine.SetTickerFactory(eng, func(time.Duration) (<-chan time.Time, func()) {
		return neverTick, func() {}
	})

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- eng.Run(ctx) }()

	cancel()

	select {
	case <-done:
		// passed
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Run did not return within 500ms after context cancel")
	}
}
