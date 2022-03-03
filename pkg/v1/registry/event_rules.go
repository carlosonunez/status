package registry

import (
	"regexp"

	v1 "github.com/carlosonunez/status/api/v1"
)

// The MatchesRegexp rule tests that an event message matches a given regexp
// pattern.
var MatchesRegexpRule v1.EventRule = v1.EventRule{
	Name: "matches-regexp",
	Evaluator: func(sut string, pattern string) (bool, error) {
		r, err := regexp.Compile(sut)
		if err != nil {
			return false, err
		}
		return r.MatchString(sut), nil
	},
}
