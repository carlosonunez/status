package pub_sub

import (
	"fmt"
	"regexp"

	v1 "github.com/carlosonunez/status/api/v1/pub_sub"
	"github.com/carlosonunez/status/third_party/registry"
)

// PubSub provides a lightweight interface for creating publish/subscriber clients and using
// them to push/pop objects into and out of the PubSub.
// See the imported API for schema information.
type PubSub interface {
	// NewPubSub creates a PubSub using the 'Credentials' and 'Properties' provided
	// and adds it to the PubSubs Client.
	NewPubSub(*v1.PubSub) error

	// GetParent returns the v1.PubSub config that created this PubSub.
	GetParent() *v1.PubSub

	// Publish publishes a message into a pub/sub topic.
	Publish(string, *v1.StatusMessage) error

	// Subscribe creates a subscription to a topic. A callback function can be
	// provided that acts on new messages sent the topic along with an arbitrary
	// set of options to provide to the underlying pub/sub implementation.
	Subscribe(string, func(m *v1.StatusMessage), interface{}) error
}

// NewPubSubFromCfg locates a PubSub given a config and returns a pointer to it.
func NewPubSubFromCfg(qCfg *v1.PubSub) (PubSub, error) {
	return locatePubSub(qCfg)
}

func locatePubSub(qCfg *v1.PubSub) (PubSub, error) {
	typeName := expandTypeName(qCfg.Type)
	t, ok := registry.IsPubSubRegistered(typeName)
	if !ok {
		return nil, fmt.Errorf("'%s' is not an implemented PubSub type", qCfg.Type)
	}
	q, ok := (t).(PubSub)
	if !ok {
		return nil, fmt.Errorf("'%s' does not implement PubSub", typeName)
	}
	if err := q.NewPubSub(qCfg); err != nil {
		return q, err
	}
	return q, nil
}

func expandTypeName(t string) string {
	re := regexp.MustCompile(`-ps$`)
	if !re.Match([]byte(t)) {
		return t + "-ps"
	}
	return t
}
