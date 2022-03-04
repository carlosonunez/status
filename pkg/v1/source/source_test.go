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

var _ = Describe("Testing events against event rules", func() {
	Context("When an event does not satisfy any rules", func() {
		It("Returns false", func() {
			evt := v1.Event{Message: "hello"}
			ruleSet := []v1.EventIncludeRule{
				{RuleType: "not-this-rule", Rule: "bar"},
				{RuleType: "not-this-rule-either", Rule: "quux"},
			}
			registry.RegisteredEventRules = []v1.EventRule{
				{
					Name: "not-this-rule",
					Evaluator: func(s string, ex string) bool {
						return false
					},
				},
				{
					Name: "not-this-rule-either",
					Evaluator: func(s string, ex string) bool {
						return false
					},
				},
			}
			defer func() { registry.RegisteredEventRules = registry.DefaultRegistryEventRules }()
			res, err := source.TestEvent(&evt, &ruleSet)
			Expect(err).To(Succeed())
			Expect(res).To(BeFalse())
		})
	})
	Context("When an event satisfies at least one event rule", func() {
		It("Returns true", func() {
			evt := v1.Event{Message: "hello"}
			ruleSet := []v1.EventIncludeRule{
				{RuleType: "not-this-rule", Rule: "bar"},
				{RuleType: "this-rule", Rule: "quux"},
			}
			registry.RegisteredEventRules = []v1.EventRule{
				{
					Name: "not-this-rule",
					Evaluator: func(s string, ex string) bool {
						return false
					},
				},
				{
					Name: "this-rule",
					Evaluator: func(s string, ex string) bool {
						return true
					},
				},
			}
			defer func() { registry.RegisteredEventRules = registry.DefaultRegistryEventRules }()
			res, err := source.TestEvent(&evt, &ruleSet)
			Expect(err).To(Succeed())
			Expect(res).To(BeTrue())
		})
	})
})

var _ = Describe("Validating Sources", func() {
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

var _ = Describe("Transforming events", func() {
	Context("When transformations are provided from a config", func() {
		Context("When a transformation completes before deadline", func() {
			It("Performs the transformations", func() {
				s := mocks.MockSource{}
				ts := []v1.EventTransform{
					{
						Input:    ".*today is (very nice)!$",
						Template: "{{ .Source.ParseDate .Message }} is {{ index .CaptureGroups 0 }}.",
					},
					{
						Input:    "(.*) is (.*)\\.$",
						Template: "{{ index .CaptureGroups 0 }} was {{ index .CaptureGroups 1 }}!",
					},
				}
				evt := v1.Event{
					Message: "Would you look at that; today is very nice!",
				}
				err := source.TransformEvent(&s, &evt, &ts)
				Expect(err).To(Succeed())
				Expect(evt.Message).To(Equal("Mar 4, 2022 was very nice!"))
			})
		})
		Context("When a transformation exceeds the deadline", func() {
			It("Errors", func() {
				source.TransformEventTimeoutSeconds = 0.5
				defer func() { source.TransformEventTimeoutSeconds = source.DefaultTransformEventTimeoutSeconds }()
				s := mocks.MockSource{}
				ts := []v1.EventTransform{
					{
						Input:    ".*today is (very nice)!$",
						Template: "{{ .Source.ReallyLongTransformFunction .Message }} is {{ index .CaptureGroups 0 }}.",
					},
					{
						Input:    "(.*) is (.*)\\.$",
						Template: "{{ index .CaptureGroups 0 }} was {{ index .CaptureGroups 1 }}!",
					},
				}
				evt := v1.Event{
					Message: "Would you look at that; today is very nice!",
				}
				err := source.TransformEvent(&s, &evt, &ts)
				Expect(err).To(MatchError("transform exceeded 500ms deadline"))
			})
		})
	})
})
