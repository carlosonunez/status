package registry

import (
	"fmt"
	"regexp"

	v1 "github.com/carlosonunez/status/api/v1"
	"github.com/carlosonunez/status/pkg/v1/interfaces"
	"github.com/carlosonunez/status/third_party/registry"
)

// RegisteredEventRules are a list of EventRules that status supports.
var RegisteredEventRules []v1.EventRule = DefaultRegistryEventRules

// DefaultRegistryEventRules is the default set of RegisteredEventRules
var DefaultRegistryEventRules []v1.EventRule = []v1.EventRule{
	MatchesRegexpRule,
}

// Type is a type of "thing" that can be registered; see the consts below.
type Type int64

const (
	// PubSub is a thing. See api/v1/types.go for more info.
	PubSub Type = iota

	// Source is a thing. See api/v1/types.go for more info.
	Source

	// Receiver is a thing. See api/v1/types.go for more info.
	Receiver

	// EventRule is a thing. See api/v1/types.go for more info.
	EventRule
)

// FindThirdPartyPubSubFn finds ThirdPartyPubSubs
var FindThirdPartyPubSubFn = registry.FindThirdPartyPubSub

// FindThirdPartySourceFn finds ThirdPartySources
var FindThirdPartySourceFn = registry.FindThirdPartySource

// LocateSource finds a third-party Source given a Source name.
func LocateSource(cfg *v1.Source) (*interfaces.Source, error) {
	t, err := locate(Source, cfg.Name)
	if err != nil {
		return nil, err
	}
	ps, ok := (t).(interfaces.Source)
	if !ok {
		return nil, fmt.Errorf("'%s' does not implement Source", cfg.Name)
	}
	return &ps, nil
}

// LocatePubSub finds a third-party PubSub given a PubSub name.
func LocatePubSub(cfg *v1.PubSub) (*interfaces.PubSub, error) {
	t, err := locate(PubSub, cfg.Type)
	if err != nil {
		return nil, err
	}
	ps, ok := (t).(interfaces.PubSub)
	if !ok {
		return nil, fmt.Errorf("'%s' does not implement PubSub", cfg.Type)
	}
	return &ps, nil
}

// LocateEventRule locates an event rule.
func LocateEventRule(r *v1.EventIncludeRule) (*v1.EventRule, error) {
	t, err := locate(EventRule, r.RuleType)
	if err != nil {
		return nil, err
	}
	rule, _ := t.(v1.EventRule)
	return &rule, nil
}

func locate(rt Type, name string) (interface{}, error) {
	var t interface{}
	var ok bool
	fullName := mustExpandName(rt, name)
	switch rt {
	case PubSub:
		t, ok = FindThirdPartyPubSubFn(fullName)
		if !ok {
			return nil, fmt.Errorf("'%s' is not an implemented PubSub", fullName)
		}
	case Source:
		t, ok = FindThirdPartySourceFn(fullName)
		if !ok {
			return nil, fmt.Errorf("'%s' is not an implemented Source", fullName)
		}
	case EventRule:
		for _, r := range RegisteredEventRules {
			if name == r.Name {
				return r, nil
			}
		}
		return nil, fmt.Errorf("'%s' is not an implement event rule", name)
	}
	return t, nil
}

func mustExpandName(t Type, name string) string {
	var suffix string

	switch t {
	case PubSub:
		suffix = "ps"
	case Source:
		suffix = "src"
	case Receiver:
		suffix = "rcv"
	}
	re := regexp.MustCompile(fmt.Sprintf("-%s$", suffix))
	if !re.Match([]byte(name)) {
		return name + "-" + suffix
	}
	return name
}
