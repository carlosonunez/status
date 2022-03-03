package pub_sub

import (
	v1 "github.com/carlosonunez/status/api/v1"
	"github.com/carlosonunez/status/pkg/v1/interfaces"
	"github.com/carlosonunez/status/pkg/v1/registry"
)

// NewPubSubFromCfg locates a PubSub given a config and returns a pointer to it.
func NewPubSubFromCfg(qCfg *v1.PubSub) (interfaces.PubSub, error) {
	t, err := registry.LocatePubSub(qCfg)
	if err != nil {
		return nil, err
	}
	return *t, nil
}
