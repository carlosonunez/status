package registry

import (
	"fmt"
	"regexp"

	v1 "github.com/carlosonunez/status/api/v1"
	"github.com/carlosonunez/status/pkg/v1/interfaces"
	"github.com/carlosonunez/status/third_party/registry"
)

type Type int64

const (
	PubSub Type = iota
	Source
	Receiver
)

var FindThirdPartyPubSubFn = registry.FindThirdPartyPubSub
var FindThirdPartySourceFn = registry.FindThirdPartySource

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
