package v1

// StatusMessage is a message that gets sent into a pub/sub queue.
type StatusMessage struct {
	// Name is the name of the message.
	Name string

	// Message is the payload containing the text that needs to be turned into a
	// status.
	Message string

	// Source is the type of source from which this StatusMessage was originated.
	Source string

	// OverrideLock determines whether this event should attempt to bypass any
	// locks applied onto the receiver.
	OverrideLock bool `json:"override_lock,omitempty"`
}
