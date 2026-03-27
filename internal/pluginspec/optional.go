package pluginspec

import "context"

// Described is an optional interface for integrations that provide a
// human-readable description, shown by "status integration show".
type Described interface {
	Description() string
}

// Sourced is an optional interface for integrations that declare their source.
// If not implemented, defaults to "builtin". External binary plugins implement
// this to return "external".
type Sourced interface {
	Source() string
}

// AuthParam describes a single authentication credential required by an
// integration, shown by "status integration show".
type AuthParam struct {
	Name        string
	Description string
	Required    bool
	EnvVar      string
}

// Authenticatable is an optional interface for integrations that require
// authentication credentials. Its metadata is shown by "status integration show".
type Authenticatable interface {
	AuthParams() []AuthParam
	AuthNotes() []string
}

// TokenStore is a minimal credential store interface used during auth flows.
// It is structurally identical to tokenstore.Store to avoid an import cycle.
type TokenStore interface {
	Get(service, key string) (string, error)
	Set(service, key, value string) error
	Delete(service, key string) error
}

// Authenticator is an optional interface for integrations that support
// interactive authentication via "status auth login".
type Authenticator interface {
	Authenticate(ctx context.Context, store TokenStore) error
}
