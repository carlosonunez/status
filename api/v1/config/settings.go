package v1

import v1 "github.com/carlosonunez/status/api/v1/pub_sub"

// Settings describe the configuration for Status itself.
type Settings struct {
	// Queue describes the queue that status should use.
	PubSub v1.PubSub `yaml:"pub_sub"`
}
