package pub_sub_test

import (
	"fmt"
	"testing"
	"time"

	v1 "github.com/carlosonunez/status/api/v1/pub_sub"
	"github.com/carlosonunez/status/pkg/v1/pub_sub"
	"github.com/carlosonunez/status/third_party/pub_sub/dummy_pubsub"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPubSubSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PubSub test suite")
}

var _ = Describe("Verifying pubSub implementation from a config", func() {
	When("A config contains an unimplemented pubSub", func() {
		It("Should produce a failure", func() {
			qCfg := v1.PubSub{
				Type: "invalid-ps",
			}
			_, err := pub_sub.NewPubSubFromCfg(&qCfg)
			Expect(err).To(MatchError("'invalid-ps' is not an implemented PubSub type"))
		})
	})
	When("A config contains an implemented pubSub that doesn't conform to PubSub", func() {
		It("Should produce a failure", func() {
			qCfg := v1.PubSub{
				Type: "nonconformant-ps",
			}
			_, err := pub_sub.NewPubSubFromCfg(&qCfg)
			Expect(err).To(MatchError("'nonconformant-ps' does not implement PubSub"))
		})
	})
	When("A config contains a valid pubSub", func() {
		It("Should return a pointer to that pubSub", func() {
			qCfg := v1.PubSub{
				Type: "example",
			}
			q, err := pub_sub.NewPubSubFromCfg(&qCfg)
			Expect(err).To(Succeed())
			err = q.NewPubSub(&qCfg)
			Expect(err).To(Succeed())
			Expect(q.GetParent()).To(Equal(&qCfg))
		})
	})
})

var _ = Describe("Publishing", func() {
	It("Can publish an event to a topic", func() {
		q, _ := pub_sub.NewPubSubFromCfg(&v1.PubSub{Type: "example"})
		err := q.Publish("topic-name", &v1.StatusMessage{
			Name:    "example-message",
			Message: "an event",
		})
		Expect(err).To(Succeed())
		Expect(q.(*dummy_pubsub.ExamplePubSub).TopicHasPendingMessages("topic-name")).To(BeTrue())
	})
})

var _ = Describe("Subscribing", func() {
	It("Can subscribe to topics", func() {
		var msg *string

		// First, let's make sure that we can subscribe to the message.
		// Since "topic-name" won't have any messages in it when we
		// call subscribe, `msg` should remain empty.
		callback := func(m *v1.StatusMessage) {
			fmt.Printf("\n===> Message: %+v", m)
			msg = &m.Message
		}
		q, _ := pub_sub.NewPubSubFromCfg(&v1.PubSub{Type: "example"})
		go func() {
			q.Subscribe("topic-name", callback, nil)
		}()
		Expect(msg).To(BeNil())

		// However, msg should get filled once we call the callback
		// upon publishing.
		_ = q.Publish("topic-name", &v1.StatusMessage{
			Name:    "example-message",
			Message: "an event",
		})

		// Publish is non-blocking, so wait a bit for msg to get allocated
		for idx := 0; idx < 500; idx++ {
			if msg != nil {
				break
			}
			time.Sleep(100)
		}
		Expect(*msg).To(Equal("an event"))
	})
})
