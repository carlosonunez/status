package registry_test

import (
	"testing"

	v1 "github.com/carlosonunez/status/api/v1"
	"github.com/carlosonunez/status/pkg/v1/pub_sub"
	"github.com/carlosonunez/status/pkg/v1/registry"
	externalregistry "github.com/carlosonunez/status/third_party/registry"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPubSubSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Locating third-party types suite")
}

var _ = Describe("Locating Pub/Subs", func() {
	When("A config contains an unimplemented pubSub", func() {
		It("Should produce a failure", func() {
			registry.FindThirdPartyPubSubFn = func(string) (interface{}, bool) {
				return nil, false
			}
			defer func() { registry.FindThirdPartyPubSubFn = externalregistry.FindThirdPartyPubSub }()
			qCfg := v1.PubSub{
				Type: "invalid-ps",
			}
			_, err := registry.LocatePubSub(&qCfg)
			Expect(err).To(MatchError("'invalid-ps' is not an implemented PubSub"))
		})
	})
	When("A config contains an implemented pubSub that doesn't implement PubSub", func() {
		It("Should produce a failure", func() {
			type InvalidPubSub struct{}
			registry.FindThirdPartyPubSubFn = func(string) (interface{}, bool) {
				ps := InvalidPubSub{}
				return &ps, true
			}
			defer func() { registry.FindThirdPartyPubSubFn = externalregistry.FindThirdPartyPubSub }()
			qCfg := v1.PubSub{
				Type: "invalid-ps",
			}
			_, err := registry.LocatePubSub(&qCfg)
			Expect(err).To(MatchError("'invalid-ps' does not implement PubSub"))
		})
	})
	When("A config contains a valid pubSub", func() {
		It("Should return a pointer to that pubSub", func() {
			// This is a real third-party pubsub in third_party.
			qCfg := v1.PubSub{
				Type: "example",
			}
			_, err := pub_sub.NewPubSubFromCfg(&qCfg)
			Expect(err).To(Succeed())
		})
	})
})
