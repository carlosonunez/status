package registry

import (
	"github.com/carlosonunez/status/third_party/pub_sub/dummy_pubsub"
	"github.com/carlosonunez/status/third_party/pub_sub/in_memory_pubsub"
)

// Registries are maps of strings to types of things to create (like pubSubs,
// sources, and receivers).
//
// Types are a list of types supported by Status that can be implemented
// externally.
//
// Ideally, I wouldn't need this and could use reflection to resolve the type
// dynamically from a string with 'unsafe' or something like that, but
// no such option exists right now.
//
// See also:
// https://stackoverflow.com/questions/23030884/is-there-a-way-to-create-an-instance-of-a-struct-from-a-string
type Type int64

const (
	PubSub Type = iota
	Source
	Receiver
)

var pubSubRegistry = map[string]interface{}{
	"nonconformant-ps": &dummy_pubsub.NonconformantPubSub{},
	"example-ps":       &dummy_pubsub.ExamplePubSub{},
	"in_memory-ps":     &in_memory_pubsub.InMemoryPubSub{},
}

func IsPubSubRegistered(n string) (interface{}, bool) {
	return isRegistered(PubSub, n)
}

// IsRegistered simply checks that a thing is registered here and returns a new
// instance of it.
func isRegistered(t Type, n string) (interface{}, bool) {
	var result bool
	var foundType interface{}
	switch t {
	case PubSub:
		foundType, result = pubSubRegistry[n]
	default:
		return nil, result
	}
	return foundType, result
}
