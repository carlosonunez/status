package fixtures

import (
	v1 "github.com/carlosonunez/status/api/v1"
	"github.com/carlosonunez/status/third_party/pub_sub/in_memory_pubsub"
)

var pubSubInstance in_memory_pubsub.InMemoryPubSub

// CreatePubSub creates a new instance of our In-Memory PubSub without
// retrieving it by name. To prevent unnecessary copies of this pub-sub in
// memory, we return a pointer to the pubSubInstance object above.
func CreatePubSub(subscriptionMap map[string]func(evt *v1.Event)) *in_memory_pubsub.InMemoryPubSub {
	pubSubInstance = in_memory_pubsub.InMemoryPubSub{}
	pubSubInstance.Initialize(&v1.PubSub{})
	for topic, callback := range subscriptionMap {
		_ = pubSubInstance.Subscribe(topic, callback, nil)
	}
	return &pubSubInstance
}
