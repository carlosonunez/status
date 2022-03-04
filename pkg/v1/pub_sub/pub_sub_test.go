package pub_sub_test

import (
	"testing"
	"time"

	v1 "github.com/carlosonunez/status/api/v1"
	"github.com/carlosonunez/status/pkg/v1/pub_sub"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPubSubSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PubSub test suite")
}

var _ = Describe("Publishing", func() {
	It("Can publish an event to a topic", func() {
		q, _ := pub_sub.NewPubSubFromCfg(&v1.PubSub{Type: "example"})
		err := q.Publish("topic-name", &v1.Event{
			Message: "an event",
		})
		Expect(err).To(Succeed())
	})
})

var _ = Describe("Subscribing", func() {
	It("Can subscribe to topics", func() {
		var msg *string

		// First, let's make sure that we can subscribe to the message.
		// Since "topic-name" won't have any messages in it when we
		// call subscribe, `msg` should remain empty.
		callback := func(m *v1.Event) {
			msg = &m.Message
		}
		q, _ := pub_sub.NewPubSubFromCfg(&v1.PubSub{Type: "example"})
		go func() {
			q.Subscribe("topic-name", callback, nil)
		}()
		Expect(msg).To(BeNil())

		// However, msg should get filled once we call the callback
		// upon publishing.
		_ = q.Publish("topic-name", &v1.Event{
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
