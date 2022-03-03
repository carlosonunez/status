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
		ps, err := pub_sub.NewPubSubFromCfg(&v1.PubSub{Type: "in_memory"})
		Expect(err).To(Succeed())
		testMap := make(map[string]string)
		cb := func(m *v1.StatusMessage) {
			if _, ok := testMap[m.Name]; !ok {
				testMap[m.Name] = m.Message
			}
		}
		msgs := []v1.StatusMessage{
			{Name: "first", Message: "first message"},
			{Name: "second", Message: "second message"},
		}
		err = ps.Subscribe("test-topic", cb, nil)
		Expect(err).To(Succeed())
		Expect(testMap).NotTo(ContainElement("first"))
		Expect(testMap).NotTo(ContainElement("second"))
		for _, msg := range msgs {
			ps.Publish("test-topic", &msg)
		}
		for idx := 0; idx < 100; idx++ {
			if _, ok := testMap["first"]; ok {
				break
			}
		}
		Expect(testMap["first"]).To(Equal("first message"))
		Expect(testMap["second"]).To(Equal("second message"))
	})
})
