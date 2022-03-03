package source_test

import (
	"testing"

	v1 "github.com/carlosonunez/status/api/v1"
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
			Expect(err).To(MatchError("invalid-source-src is not an implemented source"))
		})
	})
})
