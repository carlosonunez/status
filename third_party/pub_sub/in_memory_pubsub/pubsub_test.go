package in_memory_pubsub_test

import (
	"testing"

	v1 "github.com/carlosonunez/status/api/v1"
	"github.com/carlosonunez/status/pkg/v1/pub_sub"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestInMemoryPubSubSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Basic In-Memory PubSub Suite")
}

var _ = Describe("In-Memory PubSub", Label("component"), func() {
	It("Can publish messages and accept subscribers", func() {
		cfg := v1.PubSub{Type: "in_memory"}
		ps, err := pub_sub.NewPubSubFromCfg(&cfg)
		Expect(err).To(Succeed())
		err = ps.Initialize(&cfg)
		Expect(err).To(Succeed())
		testMap := make(map[string][]string)
		cb := func(m *v1.Event) {
			testMap["test-topic"] = append(testMap["test-topic"], m.Message)
		}
		msgs := []v1.Event{
			{Message: "first message"},
			{Message: "second message"},
		}
		err = ps.Subscribe("test-topic", cb, nil)
		Expect(err).To(Succeed())
		Expect(testMap).NotTo(ContainElement("test-topic"))
		for _, msg := range msgs {
			ps.Publish("test-topic", &msg)
		}
		// Add a small wait for the PubSub to sync
		for idx := 0; idx < 100; idx++ {
			if _, ok := testMap["first"]; ok {
				break
			}
		}
		Expect(len(testMap["test-topic"])).To(Equal(2))
		Expect(testMap["test-topic"][0]).To(Equal("first message"))
		Expect(testMap["test-topic"][1]).To(Equal("second message"))
	})
})
