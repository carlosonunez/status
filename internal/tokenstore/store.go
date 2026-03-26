// Package tokenstore provides pluggable credential storage for OAuth tokens
// used by event getters and status setters.
package tokenstore

import (
	"io/fs"
	"path/filepath"

	"github.com/adrg/xdg"
	internalconfig "github.com/carlosonunez/status/internal/config"
)

const appName = "status"

// Store persists and retrieves opaque credential strings keyed by
// (service, key). Service is e.g. "slack" or "google_calendar";
// key is e.g. "access_token" or "refresh_token".
type Store interface {
	Get(service, key string) (string, error)
	Set(service, key, value string) error
	Delete(service, key string) error
}

// FromConfig returns the Store implementation selected by cfg.
func FromConfig(cfg internalconfig.TokenStoreConfig) (Store, error) {
	switch cfg.Backend {
	case "dynamo":
		return NewDynamoStore(cfg.Dynamo)
	default: // "file" or empty string
		path := cfg.File.Path
		if path == "" {
			path = defaultFilePath()
		}
		return NewFileStore(path), nil
	}
}

func defaultFilePath() string {
	return filepath.Join(xdg.ConfigHome, appName, "tokens.json")
}

// ensure fileStore satisfies Store at compile time.
var _ Store = (*fileStore)(nil)

// ensure dynamoStore satisfies Store at compile time.
var _ Store = (*dynamoStore)(nil)

// Unexported interface used only by export_test.go to expose test constructors.
type fileReadFn = func(string) ([]byte, error)
type fileWriteFn = func(string, []byte, fs.FileMode) error
