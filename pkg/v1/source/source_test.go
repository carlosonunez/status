package source_test

import (
	"errors"
	"testing"

	v1 "github.com/carlosonunez/status/api/v1"
	"github.com/carlosonunez/status/pkg/v1/mocks"
	"github.com/carlosonunez/status/pkg/v1/registry"
	"github.com/carlosonunez/status/pkg/v1/source"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSourceSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Source reconciliation loop suite")
}

var _ = Describe("Validating event rules", func() {
	Context("When a rule has an invalid type", func() {
		It("Errors", func() {
			r := v1.EventIncludeRule{
				RuleType: "invalid",
				Rule:     "foo",
			}
			registry.RegisteredEventRules = []v1.EventRule{}
			defer func() { registry.RegisteredEventRules = registry.DefaultRegistryEventRules }()
			err := source.ValidateEventRule(&r)
			Expect(err).NotTo(Succeed())
		})
	})
})

var _ = Describe("Validating", func() {
	Context("When a source has an event with an invalid rule", func() {
		It("Errors", func() {
			source.ValidateEventRuleFn = func(r *v1.EventIncludeRule) error {
				if r.RuleType == "invalid-rule" {
					return errors.New("an error")
				}
				return nil
			}
			defer func() { source.ValidateEventRuleFn = source.ValidateEventRule }()
			err := source.ValidateSource(&mocks.SourceWithInvalidEventRules)
			Expect(err).To(MatchError("'invalid-rule' in event 'valid-event-2' in " +
				"source 'invalid-source' is an invalid event rule: 'an error'"))
		})
	})
})

var _ = Describe("Resolving sources from a config", func() {
	Context("When a given name doesn't have a registered source", func() {
		It("produces a failure", func() {
			_, err := source.NewSourceFromCfg(&v1.Source{
				Name: "invalid-source",
			})
			Expect(err).To(MatchError("'invalid-source-src' is not an implemented Source"))
		})
	})
})
