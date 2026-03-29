package pluginsdk

import "time"

// Event represents a calendar or scheduling event returned by a getter plugin.
type Event struct {
	Calendar string         `json:"calendar"`
	Title    string         `json:"title"`
	StartsAt time.Time      `json:"starts_at"`
	EndsAt   time.Time      `json:"ends_at"`
	AllDay   bool           `json:"all_day"`
	Meta     map[string]any `json:"meta,omitempty"`
}

// ParamSpec describes a single parameter accepted by a plugin.
type ParamSpec struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
	Type        string `json:"type"`
}

// SetResult is the outcome of a SetStatus call.
type SetResult struct {
	Skipped  bool   `json:"skipped"`
	Response string `json:"response,omitempty"`
}

// --- Internal protocol messages (unexported; shared format with internal/plugin) ---

type metadataResponse struct {
	Name         string      `json:"name"`
	Type         string      `json:"type"` // "getter" or "setter"
	MinInterval  string      `json:"min_interval,omitempty"`
	ParamSpecs   []ParamSpec `json:"param_specs,omitempty"`
	SupportsAuth bool        `json:"supports_auth,omitempty"`
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
	Events []Event `json:"events"`
	Error  string  `json:"error,omitempty"`
}

type setStatusRequest struct {
	Params map[string]any `json:"params"`
}

type setStatusResponse struct {
	Skipped  bool   `json:"skipped"`
	Response string `json:"response,omitempty"`
	Error    string `json:"error,omitempty"`
}
