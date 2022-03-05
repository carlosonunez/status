package interfaces

import v1 "github.com/carlosonunez/status/api/v1"

// PubSub provides a lightweight interface for creating publish/subscriber clients and using
// them to push/pop objects into and out of the PubSub.
// See the imported API for schema information.
type PubSub interface {
	// Initialize creates a PubSub using the 'Credentials' and 'Properties' provided
	// and adds it to the PubSubs Client.
	Initialize(*v1.PubSub) error

	// GetParent returns the v1.PubSub config that created this PubSub.
	GetParent() *v1.PubSub

	// Publish publishes a message into a pub/sub topic.
	Publish(string, *v1.Event) error

	// Subscribe creates a subscription to a topic. A callback function can be
	// provided that acts on new messages sent the topic along with an arbitrary
	// set of options to provide to the underlying pub/sub implementation.
	Subscribe(string, func(m *v1.Event), interface{}) error
}

// Source is a generator of events that might produce statuses.
type Source interface {
	// Initialize initializes your source. Configuration properties are defined in
	// status.yaml and sent as a string/interface{} map.
	//
	// Use this to create clients, log into services, etc.
	Initialize(map[string]interface{}) error

	// GetParent retrieves the config that spawnwed the Source.
	GetParent() *v1.Source

	// Poll retrieves events from your source at the interval specified by
	// poll_interval.
	//
	// CoC: Do not use Poll to process/filter your events! Status will take care
	// of that for you.
	Poll() (*[]*v1.Event, error)
}
