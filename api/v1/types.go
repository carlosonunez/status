package v1

// Config describes the schema for configuring Status.
type Config struct {

	// Sources are a list of event sources that produce statuses.
	Sources []Source

	// Receivers are a list of receivers onto which statuses are posted.
	Receivers interface{}

	// Settings describe the configuration for Status itself.
	Settings Settings
}

// Settings describe the configuration for Status itself.
type Settings struct {
	// Queue describes the queue that status should use.
	PubSub PubSub `yaml:"pub_sub"`
}

// Source is a service or platform that generates events that should turn into
// statuses.
// Google Calendar, RSS, or GitHub are some examples of potential
// `Sources`.
//
// Sources send "events". Events are things that should turn into statuses, like
// calendar entries, feed entries, or labels on a GitHub issue.
//
// Events can "lock" a receiver (see receiver.go) if they want the status
// generated by this event to persist for a period of time. Note that other
// events can override receiver locks depending on their sources's relative
// weights.
type Source struct {
	// Name is the name of the source. Must be unique.
	Name string

	// Settings is a configuration block for this source.
	Settings SourceSettings

	// Properties are an arbitrary map of settings to provide to this source.
	// Credentials, URLs, and service-specific settings are good examples of
	// Properties to set.
	Properties map[string]interface{} `yaml:",omitempty"`

	// StatusGeneratingEvents are events proudced by the source's parent service that should
	// produce a status.
	// Use Events to filter what a source produces.
	StatusGeneratingEvents []StatusGeneratingEvent `yaml:"events"`
}

// SourceSettings defines sources.
type SourceSettings struct {
	// Enabled determines whether this status should enable this source (i.e. poll
	// its service). 'true' by default.
	Enabled bool

	// LockDurations are a list of periods that events generated by this source
	// should lock their receivers by default. Events can override them as needed.
	LockDurations []LockDuration `yaml:"lock_durations:omitempty"`

	// Weight is the rank of this Sources.
	// Only one status can be set per receiver per lock duration. The source with
	// the lowest weight wins. No two weights can be the same.
	Weight int

	// PollInterval determines how often Status should poll the source's service
	// in minutes. Status polls every five (5) minutes by default.
	PollDuration int `yaml:"poll_interval,omitempty"`
}

// StatusGeneratingEvent is an event that produces a status.
type StatusGeneratingEvent struct {
	// Name is the name of an event. Must be unique.
	Name string

	// Enabled determines whether status should act on this event. Defaults to
	// 'true'.
	Enabled bool `yaml:"omitempty"`

	// Weight is the weight of this event. Highest event with payloads wins.
	Weight int

	// IncludeIf are a list of conditions to watch for.
	IncludeIf []EventIncludeRule

	// AlwaysOverwrite will force the receiver to write the status created by this
	// event.
	AlwaysOverwrite bool `yaml:"always_overwrite,omitempty"`

	// Transforms are a list of transformations to apply onto the payload in this
	// event to produce a status.
	// If omitted, the payload in the event is sent as-is.
	Transforms []EventTransform `yaml:",omitempty"`

	// LockDuration is a period of time for which events generated by this status
	// should be "locked" to their receivers.
	// Example: A 30-minute calendar meeting might post a status to all receivers
	// that can't be changed for 30 minutes.
	LockDuration LockDuration `yaml:"lock_duration,omitempty"`
}

// LockDuration is a period of time for which events generated by this status
// should be "locked" to their receivers.
// Example: A 30-minute calendar meeting might post a status to all receivers
// that can't be changed for 30 minutes.
type LockDuration struct {
	// DefaultDuration is the length of the duration in minutes. This applies for ALL receivers.
	DefaultDuration int `yaml:"default,omitempty"`

	// Exceptions are durations applied onto specific receivers by name.
	Exceptions []LockDurationException `yaml:",omitempty"`
}

// LockDurationException is a lock duration for a specific receiver.
type LockDurationException struct {
	// ReceiverName is the name of the receiver for whom this exception applies.
	ReceiverName string `yaml:"receiver"`

	// Duration is the length of the duration in minutes for this receiver.
	Duration int
}

// EventIncludeRule is a rule to evaluate to determine if the event should
// produce a status.
type EventIncludeRule struct {
	// RuleType is the type of rule to evaluate. See pkg/v1/event/event_rule.go
	// for more details.
	RuleType string

	// Rule is a string to evaluate against the event payload.
	Rule string
}

// EventTransform is a transform to apply onto events that satsify their
// EventRules.
type EventTransform struct {
	// Input is a regexp to use to capture information from the payload.
	Input string

	// Template is a go-template to apply on top of the payload
	Template string
}

// PubSub is an abstract client for publish-subscriber services.
// Its goal is to provide a lightweight way of instantiating pub/subs and
// pushing into/popping from them.
type PubSub struct {
	// Type is the type of the pub/sub.
	// PubSubs are implemented in 'third_party/pub_sub'.
	Type string

	// Credentials holds authentication information for creating the pub/sub.
	Credentials map[string]interface{}

	// Properties holds additional configuration information needed to construct
	// the pub/sub, like delays and read/write options.
	Properties map[string]interface{}
}

// EventRule is a rule evaluated against EventIncludeRules.
type EventRule struct {
	// Name is the name of the rule. This is provided to EventIncludeRules in the
	// 'Rule' field.
	Name string

	// Evaluator is a function that the event payload is tested against.
	// The event payload is sent as the first argument, and the condition inside
	// of the EventIncludeRule is sent as the second. A context is provided to the
	// Evaluator in case you want to send a hint back afte evaluation.
	//
	// Both are assumed to be strings for now.
	Evaluator func(string, string) bool
}

// Event is an event generated by a Source.
type Event struct {
	// Message is a payload of text that should be turned into a status.
	Message string

	// Targets are a list of receivers (by name) onto which the status should be posted.
	// Empty slices imply that the status should be applied onto all receivers.
	Targets []string
}
