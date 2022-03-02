package v1

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
