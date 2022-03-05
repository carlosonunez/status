package source_test

import (
	"time"

	v1 "github.com/carlosonunez/status/api/v1"
	"github.com/carlosonunez/status/pkg/v1/fixtures"
	"github.com/carlosonunez/status/pkg/v1/source"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Polling", Label("integration"), func() {
	When("The source can poll successfully", func() {
		It("Submits transformed events into a registered queue", func() {
			src := fixtures.ExampleSource{}
			strs := []string{}
			p := &strs
			subscriptionCallbacks := map[string]func(evt *v1.Event){
				"all-receivers": func(evt *v1.Event) {
					*p = append(*p, evt.Message)
				},
			}
			pubSub := fixtures.CreatePubSub(subscriptionCallbacks)
			err := source.FetchAndPublishEvents(&src, pubSub)
			for idx := 0; idx < 20; idx++ {
				if len(strs) == 1 {
					break
				}
				time.Sleep(250 * time.Millisecond)
			}
			Expect(err).To(Succeed())
			Expect(len(strs)).To(Equal(1))
			Expect(strs[0]).To(Equal("My event happened on Mar 4, 2022"))
		})
	})
})
