package transform

import (
	"fmt"
	"regexp"

	"github.com/carlosonunez/status/internal/getter"
	"github.com/carlosonunez/status/internal/params"
	"github.com/carlosonunez/status/internal/setter"
)

// StatusTemplate holds the un-resolved parameters for one setter within a
// transform. String values in Params may contain $N references that are
// substituted with regex capture groups when the transform is applied.
type StatusTemplate struct {
	Params params.Params
}

// Transform matches an event title against a regex pattern and resolves status
// parameters for one or more setters.
type Transform struct {
	Name    string
	pattern *regexp.Regexp
	setters map[string]StatusTemplate
}

// New compiles pattern and returns a Transform. Returns an error if pattern is
// not a valid regular expression.
func New(pattern string, setters map[string]StatusTemplate) (*Transform, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid pattern %q: %w", pattern, err)
	}
	return &Transform{
		pattern: re,
		setters: setters,
	}, nil
}

// Apply attempts to match event.Title against the transform's pattern.
// If the pattern matches and setterName is configured on this transform, it
// returns a setter.Status with all string param values resolved ($N references
// substituted) and true. Returns a zero setter.Status and false otherwise.
func (t *Transform) Apply(event getter.Event, setterName string) (setter.Status, bool) {
	tmpl, ok := t.setters[setterName]
	if !ok {
		return setter.Status{}, false
	}
	if !t.pattern.MatchString(event.Title) {
		return setter.Status{}, false
	}

	resolved := make(map[string]any, len(tmpl.Params.Keys()))
	for _, k := range tmpl.Params.Keys() {
		if s, ok := tmpl.Params.String(k); ok {
			resolved[k] = t.substituteCaptures(s, event.Title)
		} else if v, ok := tmpl.Params.Value(k); ok {
			resolved[k] = v
		}
	}
	return setter.Status{Params: params.FromMap(resolved)}, true
}

// SetterNames returns the names of all setters configured on this transform.
func (t *Transform) SetterNames() []string {
	names := make([]string, 0, len(t.setters))
	for name := range t.setters {
		names = append(names, name)
	}
	return names
}

// substituteCaptures substitutes $N references in tpl with capture groups from
// matching title against the transform's pattern.
// $N is normalised to ${N} before substitution to avoid ambiguity when $N is
// immediately adjacent to alphanumeric characters (e.g. "$1foo").
func (t *Transform) substituteCaptures(tpl, title string) string {
	return t.pattern.ReplaceAllString(title, normalizeDollarN(tpl))
}

var dollarN = regexp.MustCompile(`\$(\d+)`)

func normalizeDollarN(s string) string {
	return dollarN.ReplaceAllString(s, "$${${1}}")
}
