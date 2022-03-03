package source

import (
	v1 "github.com/carlosonunez/status/api/v1"
	"github.com/carlosonunez/status/pkg/v1/interfaces"
	"github.com/carlosonunez/status/pkg/v1/registry"
)

// NewSourceFromCfg creates an instance of a source.
func NewSourceFromCfg(cfg *v1.Source) (*interfaces.Source, error) {
	s, err := registry.LocateSource(cfg)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// ValidateSource checks that a source defined in a config is correct.
// This is done during status initialization before the source poll loop.
// We fail hard if a source is not defined to prevent statuses not getting set
// due to invalid configs.
func ValidateSource(s *v1.Source) error {
	for _, e := range s.StatusGeneratingEvents {
		for _, r := range e.IncludeIf {
			if err := ValidateEventRuleFn(&r); err != nil {
				return fmt.Errorf("'%s' in event '%s' in source '%s' is an invalid event rule: '%s'",
					r.RuleType, e.Name, s.Name, err)
			}
		}
	}
	return nil
}

// ValidateEventRule validates an event rule.
func ValidateEventRule(r *v1.EventIncludeRule) error {
	for _, er := range registry.RegisteredEventRules {
		if strings.ToLower(r.Rule) == strings.ToLower(er.Name) {
			return nil
		}
	}
	return fmt.Errorf("'%s' does not have an accompanying registered event rule", r.Rule)
}
