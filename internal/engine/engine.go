// Package engine drives the event-getter → transform → status-setter pipeline.
// It polls registered event getters on a schedule, applies matching transforms,
// and dispatches resolved statuses to the appropriate setters.
package engine

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/carlosonunez/status/internal/config"
	"github.com/carlosonunez/status/internal/getter"
	"github.com/carlosonunez/status/internal/params"
	"github.com/carlosonunez/status/internal/rule"
	"github.com/carlosonunez/status/internal/setter"
	"github.com/carlosonunez/status/internal/transform"
	"github.com/sirupsen/logrus"
)

// getterGroup bundles a single EventGetter with all rules that use it,
// sorted by ascending priority. Multiple rules may share a getter; when they
// do, they must share the same getter params (the first rule's params are used).
type getterGroup struct {
	getter getter.EventGetter
	params params.Params
	rules  []*rule.Rule
}

// Engine drives the pipeline.
type Engine struct {
	groups    []getterGroup
	setters   *setter.Registry
	now       func() time.Time
	newTicker func(time.Duration) (<-chan time.Time, func())
}

// New validates that every getter and setter name referenced in cfg is
// registered, then returns a ready Engine.
func New(cfg *config.Config, gr *getter.Registry, sr *setter.Registry) (*Engine, error) {
	rules := make([]*rule.Rule, 0, len(cfg.Rules))
	for _, rc := range cfg.Rules {
		r, err := rule.FromConfig(rc)
		if err != nil {
			return nil, fmt.Errorf("engine: %w", err)
		}
		rules = append(rules, r)
	}
	rule.SortByPriority(rules)

	// Group rules by getter name; validate getter names; track insertion order.
	type entry struct {
		getter getter.EventGetter
		params params.Params
		rules  []*rule.Rule
	}
	var order []string
	byName := make(map[string]*entry)

	for _, r := range rules {
		if _, exists := byName[r.GetterName]; !exists {
			g, err := gr.Get(r.GetterName)
			if err != nil {
				return nil, fmt.Errorf("engine: rule %q references unknown getter %q", r.Name, r.GetterName)
			}
			byName[r.GetterName] = &entry{
				getter: g,
				params: params.FromMap(r.GetterParams),
			}
			order = append(order, r.GetterName)
		}
		byName[r.GetterName].rules = append(byName[r.GetterName].rules, r)
	}

	// Validate all setter names across every rule's transforms.
	for _, r := range rules {
		for _, name := range allSetterNames(r.Transforms) {
			if _, err := sr.Get(name); err != nil {
				return nil, fmt.Errorf("engine: rule %q references unknown setter %q", r.Name, name)
			}
		}
	}

	groups := make([]getterGroup, 0, len(order))
	for _, name := range order {
		e := byName[name]
		groups = append(groups, getterGroup{getter: e.getter, params: e.params, rules: e.rules})
	}

	return &Engine{
		groups:    groups,
		setters:   sr,
		now:       time.Now,
		newTicker: defaultTicker,
	}, nil
}

// Run starts one polling goroutine per getter group and blocks until ctx is
// cancelled. All goroutines are waited for before returning.
func (e *Engine) Run(ctx context.Context) error {
	var wg sync.WaitGroup
	for _, g := range e.groups {
		g := g
		wg.Add(1)
		go func() {
			defer wg.Done()
			e.runGroup(ctx, g)
		}()
	}
	wg.Wait()
	return nil
}

func (e *Engine) runGroup(ctx context.Context, g getterGroup) {
	ticks, stop := e.newTicker(g.getter.MinInterval())
	defer stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticks:
			e.processGroup(ctx, g)
		}
	}
}

// processGroup fetches events for one getter group, filters to active events,
// then dispatches statuses in rule-priority order.
func (e *Engine) processGroup(ctx context.Context, g getterGroup) {
	events, err := g.getter.GetEvents(ctx, g.params)
	if err != nil {
		logrus.WithError(err).WithField("getter", g.getter.Name()).Warn("getter error; skipping tick")
		return
	}
	for _, event := range events {
		if !isActive(event, e.now()) {
			continue
		}
		dispatched := make(map[string]bool)
		for _, r := range g.rules {
			e.dispatch(ctx, event, r, dispatched)
		}
	}
}

// dispatch iterates a rule's transforms and calls SetStatus on each setter
// whose pattern first matches the event. A setter is called at most once per
// event (tracked by dispatched).
func (e *Engine) dispatch(ctx context.Context, event getter.Event, r *rule.Rule, dispatched map[string]bool) {
	for _, setterName := range allSetterNames(r.Transforms) {
		if dispatched[setterName] {
			continue
		}
		s, err := e.setters.Get(setterName)
		if err != nil {
			logrus.WithField("setter", setterName).WithField("rule", r.Name).Warn("setter not found")
			continue
		}
		for _, tr := range r.Transforms {
			status, ok := tr.Apply(event, setterName)
			if !ok {
				continue
			}
			if _, err := s.SetStatus(ctx, status); err != nil {
				logrus.WithError(err).WithField("setter", setterName).WithField("rule", r.Name).Warn("setter error")
			}
			dispatched[setterName] = true
			break
		}
	}
}

// isActive reports whether event is currently happening at now.
func isActive(event getter.Event, now time.Time) bool {
	if event.AllDay {
		today := now.UTC().Truncate(24 * time.Hour)
		day := event.StartsAt.UTC().Truncate(24 * time.Hour)
		return today.Equal(day)
	}
	return !now.Before(event.StartsAt) && now.Before(event.EndsAt)
}

// allSetterNames returns a deduplicated list of all setter names referenced
// across the given transforms, in first-seen order.
func allSetterNames(transforms []*transform.Transform) []string {
	seen := make(map[string]bool)
	var names []string
	for _, tr := range transforms {
		for _, name := range tr.SetterNames() {
			if !seen[name] {
				seen[name] = true
				names = append(names, name)
			}
		}
	}
	return names
}

func defaultTicker(d time.Duration) (<-chan time.Time, func()) {
	t := time.NewTicker(d)
	return t.C, t.Stop
}
