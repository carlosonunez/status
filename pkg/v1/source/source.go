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
