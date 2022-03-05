package config_test

import (
	"io/ioutil"

	"github.com/carlosonunez/status/pkg/v1/config"
	"github.com/carlosonunez/status/pkg/v1/mocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/r3labs/diff/v2"
)

var _ = Describe("Reading config files", func() {
	When("Config file invalid", func() {
		It("Fails to generate a config", func() {
			yaml := `
---
invalid_key: invalid_yaml
`
			_, err := config.Parse(yaml)
			Expect(err).To(MatchError("config file is invalid: line 3: field invalid_key not found in type v1.Config"))
		})
	})
	When("Config file valid", func() {
		It("Generates a config", func() {
			s, err := ioutil.ReadFile("../mocks/status.yaml")
			if err != nil {
				panic(err)
			}
			cfg, err := config.Parse(string(s))
			Expect(err).To(Succeed())
			changelog, _ := diff.Diff(*cfg, mocks.RenderedConfig)
			if len(changelog) > 0 {
				Expect(diff.Diff(*cfg, mocks.RenderedConfig)).To(Equal(""))
			} else {
				Expect(len(changelog)).To(Equal(0))
			}
		})
	})
})
