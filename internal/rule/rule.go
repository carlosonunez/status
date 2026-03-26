package rule

import (
	"fmt"
	"sort"

	"github.com/carlosonunez/status/internal/config"
	"github.com/carlosonunez/status/internal/transform"
)

// Rule is a parsed, ready-to-evaluate rule connecting an EventGetter to
// status setters through a list of ordered transforms.
type Rule struct {
	Name         string
	Priority     int
	GetterName   string
	GetterParams map[string]any
	Transforms   []*transform.Transform
}

// FromConfig parses a config.RuleConfig into a Rule, compiling all transform
// patterns. Returns an error if any pattern is invalid.
func FromConfig(cfg config.RuleConfig) (*Rule, error) {
	transforms := make([]*transform.Transform, 0, len(cfg.Transforms))

	for _, tc := range cfg.Transforms {
		setters := make(map[string]transform.StatusTemplate, len(tc.StatusSetters))
		for _, ss := range tc.StatusSetters {
			params := ss.Params
			if params == nil {
				params = map[string]any{}
			}
			setters[ss.Name] = transform.StatusTemplate{Params: params}
		}

		tr, err := transform.New(tc.Pattern, setters)
		if err != nil {
			return nil, fmt.Errorf("rule %q, transform %q: %w", cfg.Name, tc.Name, err)
		}
		tr.Name = tc.Name
		transforms = append(transforms, tr)
	}

	return &Rule{
		Name:         cfg.Name,
		Priority:     cfg.Priority,
		GetterName:   cfg.EventGetter.Name,
		GetterParams: cfg.EventGetter.Params,
		Transforms:   transforms,
	}, nil
}

// SortByPriority sorts rules ascending by Priority. Rules with Priority 0
// (unset) are placed after all explicitly-prioritised rules.
func SortByPriority(rules []*Rule) {
	sort.SliceStable(rules, func(i, j int) bool {
		pi, pj := rules[i].Priority, rules[j].Priority
		if pi == 0 {
			return false
		}
		if pj == 0 {
			return true
		}
		return pi < pj
	})
}
