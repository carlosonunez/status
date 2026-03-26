package transform

import (
	"fmt"
	"regexp"
	"time"

	"github.com/carlosonunez/status/internal/getter"
)

// StatusTemplate holds the un-resolved status fields for one setter within a transform.
// MessageTpl may contain $N references that are replaced with regex capture groups.
type StatusTemplate struct {
	MessageTpl    string
	Emoji         string
	Duration      *time.Duration
	IsOutOfOffice bool
}

// ResolvedStatus is a StatusTemplate with all $N references substituted.
type ResolvedStatus struct {
	Message       string
	Emoji         string
	Duration      *time.Duration
	IsOutOfOffice bool
}

// Transform matches an event title against a regex pattern and resolves status
// templates for one or more setters.
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
// If the pattern matches and setterName is configured on this transform,
// it returns the resolved ResolvedStatus and true.
// Returns a zero ResolvedStatus and false otherwise.
func (t *Transform) Apply(event getter.Event, setterName string) (ResolvedStatus, bool) {
	tmpl, ok := t.setters[setterName]
	if !ok {
		return ResolvedStatus{}, false
	}

	if !t.pattern.MatchString(event.Title) {
		return ResolvedStatus{}, false
	}

	return ResolvedStatus{
		Message:       t.resolveTemplate(tmpl.MessageTpl, event.Title),
		Emoji:         tmpl.Emoji,
		Duration:      tmpl.Duration,
		IsOutOfOffice: tmpl.IsOutOfOffice,
	}, true
}

// resolveTemplate substitutes $N references in tpl with the capture groups
// from matching title against the transform's pattern.
// $N is normalised to ${N} before substitution to avoid ambiguity when $N
// is adjacent to alphanumeric characters.
func (t *Transform) resolveTemplate(tpl, title string) string {
	normalised := normalizeDollarN(tpl)
	return t.pattern.ReplaceAllString(title, normalised)
}

// normalizeDollarN converts bare $N references (e.g. $1, $12) to the
// unambiguous ${N} form required by regexp.ReplaceAllString.
var dollarN = regexp.MustCompile(`\$(\d+)`)

func normalizeDollarN(s string) string {
	return dollarN.ReplaceAllString(s, "$${${1}}")
}
