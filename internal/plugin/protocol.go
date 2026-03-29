// Package plugin handles external binary plugin discovery and subprocess
// communication for status-getter-<name> and status-setter-<name> binaries.
package plugin

import "github.com/carlosonunez/status/internal/pluginspec"

// These types mirror pkg/pluginsdk's unexported protocol messages.
// They live here so internal/plugin can decode responses from external plugins
// without importing the public SDK package.

type metadataResponse struct {
	Name         string                `json:"name"`
	Type         string                `json:"type"` // "getter" or "setter"
	MinInterval  string                `json:"min_interval,omitempty"`
	ParamSpecs   []pluginspec.ParamSpec `json:"param_specs,omitempty"`
	SupportsAuth bool                  `json:"supports_auth,omitempty"`
}

type authenticateRequest struct {
	Params map[string]any `json:"params"`
}

type authenticateResponse struct {
	Tokens map[string]string `json:"tokens,omitempty"`
	Error  string            `json:"error,omitempty"`
}

type getEventsRequest struct {
	Params map[string]any `json:"params"`
}

type getEventsResponse struct {
	Events []externalEvent `json:"events"`
	Error  string          `json:"error,omitempty"`
}

type externalEvent struct {
	Calendar string         `json:"calendar"`
	Title    string         `json:"title"`
	StartsAt string         `json:"starts_at"`
	EndsAt   string         `json:"ends_at"`
	AllDay   bool           `json:"all_day"`
	Meta     map[string]any `json:"meta,omitempty"`
}

type setStatusRequest struct {
	Params map[string]any `json:"params"`
}

type setStatusResponse struct {
	Skipped  bool   `json:"skipped"`
	Response string `json:"response,omitempty"`
	Error    string `json:"error,omitempty"`
}
