package registry

import (
	"github.com/carlosonunez/status/third_party/pub_sub/dummy_pubsub"
	"github.com/carlosonunez/status/third_party/pub_sub/in_memory_pubsub"
)

// This is the registry of PubSubs, sources, and receivers.
// You can override it during testing if needed; please restore it using
// a defer() if you do!
var ThirdPartyRegistry map[string]map[string]interface{} = registry

var registry = map[string]map[string]interface{}{
	"pubsub": {
		"nonconformant-ps": &dummy_pubsub.NonconformantPubSub{},
		"example-ps":       &dummy_pubsub.ExamplePubSub{},
		"in_memory-ps":     &in_memory_pubsub.InMemoryPubSub{},
	},
	"source": {},
}

// FindThirdPartyPubSub locates a registered pubsub.
func FindThirdPartyPubSub(n string) (interface{}, bool) {
	return locate(n, "pubsub")
}

// FindThirdPartySource locates registered sources
func FindThirdPartySource(n string) (interface{}, bool) {
	return locate(n, "source")
}

func locate(n string, k string) (interface{}, bool) {
	t, ok := ThirdPartyRegistry[k][n]
	if !ok {
		return nil, false
	}
	return t, true
}
