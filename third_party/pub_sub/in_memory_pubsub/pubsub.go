package in_memory_pubsub

import (
	"fmt"

	v1 "github.com/carlosonunez/status/api/v1"
	"github.com/simonfxr/pubsub"
)

// InMemoryPubSub is a Pub/Sub that works entirely in memory. This is a good
// option for running Status in dev mode on your machine.
// Use one of the "cloud-y" Pub/Subs if you're aiming to implement Status in
// production.
type InMemoryPubSub struct {
	// Config is a pointer ot the config that spawned it.
	Config *v1.PubSub

	// Client is the underlying client providing the pub/sub.
	Client *pubsub.Bus

	// SubscriptionMap contains a list of subscriptions open per topic.
	SubscriptionMap map[string]*pubsub.Subscription
}

// InMemoryPubSubProperties defines how props should be laid out for this
// pub/sub.
type InMemoryPubSubProperties struct {
	QueueSize  int `yaml:"queue_size:omitempty"`
	TopicCount int `yaml:"topic_count:omitempty"`
}

// Initialize
func (c *InMemoryPubSub) Initialize(cfg *v1.PubSub) error {
	c.Client = pubsub.NewBus()
	c.Config = cfg
	c.SubscriptionMap = make(map[string]*pubsub.Subscription)
	return nil
}

// GetParent
func (c *InMemoryPubSub) GetParent() *v1.PubSub {
	return c.Config
}

// Subscribe
func (c *InMemoryPubSub) Subscribe(topic string, cb func(e *v1.Event), opts interface{}) error {
	if _, ok := c.SubscriptionMap[topic]; ok {
		return fmt.Errorf("a subscription is already open for topic '%s'", topic)
	}
	sub := c.Client.Subscribe(topic, cb)
	c.SubscriptionMap[topic] = sub
	return nil
}

// Publish
func (c *InMemoryPubSub) Publish(topic string, m *v1.Event) error {
	c.Client.Publish(topic, m)
	return nil
}
