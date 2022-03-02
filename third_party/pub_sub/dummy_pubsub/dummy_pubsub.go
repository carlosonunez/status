package dummy_pubsub

import (
	"fmt"

	v1 "github.com/carlosonunez/status/api/v1/pub_sub"
)

// These are queues used for testing.

// NonconformantPubSub is a queue that doesn't implement the PubSub interface.
type NonconformantPubSub struct{}

// ExamplePubSub is an example of a valid queue.
type ExamplePubSub struct {
	// Config is a pointer to the configuration that spawned it.
	Config *v1.PubSub

	// Name is the name of the example queue.
	Name string

	// Client is an underlying "client" implementing a pub/sub.
	Client examplePubSubClient
}

// Right, so this is not at all a real implementation of a pub/sub. This is
// meant to confirm that the underlying abstractions are working and that
// everything is connected.
type examplePubSubClient struct {
	// Topics gets incremented when a new topic is created.
	Topics int

	// Messages is a map of topics to message queues inside of them.
	Messages map[string]chan *v1.StatusMessage
}

// NewPubSub creates a new example queue
func (c *ExamplePubSub) NewPubSub(qCfg *v1.PubSub) error {
	c.Config = qCfg
	return nil
}

// GetParent returns the config.
func (c *ExamplePubSub) GetParent() *v1.PubSub {
	return c.Config
}

// Publish publishes into a dummy topic queue.
func (c *ExamplePubSub) Publish(topic string, msg *v1.StatusMessage) error {
	if c.Client.Messages == nil {
		c.Client.Messages = make(map[string]chan *v1.StatusMessage)
	}
	if _, ok := c.Client.Messages[topic]; !ok {
		c.Client.Messages[topic] = make(chan *v1.StatusMessage, 1)
	}
	select {
	case c.Client.Messages[topic] <- msg:
		return nil
	default:
		return fmt.Errorf("message not sent to '%s': %+v", topic, msg)
	}
}

// Subscribe creates a subscription to a topic.
// In this case, it simply confirms that subscribe is called eventually when
// executed from a goroutine.
func (c *ExamplePubSub) Subscribe(topic string, f func(m *v1.StatusMessage), opts interface{}) error {
	f(<-c.Client.Messages[topic])
	return nil
}

// TopicHasPendingMessages determines if the topic has any messages that have
// not been claimed yet.
func (c *ExamplePubSub) TopicHasPendingMessages(topic string) bool {
	if _, ok := c.Client.Messages[topic]; !ok {
		return false
	}
	return len(c.Client.Messages[topic]) > 0
}

// Clear purges all messages from all topics.
func (c *ExamplePubSub) Clear() {
	for _, t := range c.Client.Messages {
		close(t)
	}
}
