// Package pluginsdk provides the types and helpers needed to implement
// external status-getter-<name> and status-setter-<name> plugin binaries.
// Plugin authors should import this package and call ServeGetter or ServeSetter
// from their main function, then exit with the returned code.
package pluginsdk

import (
	"encoding/json"
	"io"
	"os"
)

// GetterMetadata describes a getter plugin returned by its --metadata flag.
type GetterMetadata struct {
	Name         string      `json:"name"`
	MinInterval  string      `json:"min_interval,omitempty"`
	ParamSpecs   []ParamSpec `json:"param_specs,omitempty"`
	SupportsAuth bool        `json:"supports_auth,omitempty"`
}

// SetterMetadata describes a setter plugin returned by its --metadata flag.
type SetterMetadata struct {
	Name         string      `json:"name"`
	ParamSpecs   []ParamSpec `json:"param_specs,omitempty"`
	SupportsAuth bool        `json:"supports_auth,omitempty"`
}

// GetterHandler is implemented by external getter plugins.
type GetterHandler interface {
	Metadata() GetterMetadata
	GetEvents(params map[string]any) ([]Event, error)
}

// SetterHandler is implemented by external setter plugins.
type SetterHandler interface {
	Metadata() SetterMetadata
	SetStatus(params map[string]any) (SetResult, error)
}

// AuthHandler is an optional interface implemented by plugins that support
// interactive authentication. ServeGetter and ServeSetter will handle
// --authenticate automatically when the handler implements this interface.
type AuthHandler interface {
	// Authenticate runs the authentication flow, writing user-facing prompts
	// and URLs to out. It returns the acquired tokens as a name→value map
	// (e.g. {"access_token": "…", "token_secret": "…"}).
	Authenticate(params map[string]any, out io.Writer) (map[string]string, error)
}

// ServeGetter dispatches a single plugin invocation for a getter binary.
// args should be os.Args[1:], in should be os.Stdin, out should be os.Stdout.
// Returns 0 on success, 1 on error. Plugin binaries should os.Exit with this value.
func ServeGetter(h GetterHandler, args []string, in io.Reader, out io.Writer) int {
	if len(args) == 0 {
		return 1
	}
	enc := json.NewEncoder(out)
	switch args[0] {
	case "--metadata":
		m := h.Metadata()
		resp := metadataResponse{
			Name:         m.Name,
			Type:         "getter",
			MinInterval:  m.MinInterval,
			ParamSpecs:   m.ParamSpecs,
			SupportsAuth: m.SupportsAuth,
		}
		if _, ok := h.(AuthHandler); ok {
			resp.SupportsAuth = true
		}
		if err := enc.Encode(resp); err != nil {
			return 1
		}
		return 0

	case "--get-events":
		var req getEventsRequest
		if err := json.NewDecoder(in).Decode(&req); err != nil {
			return 1
		}
		events, err := h.GetEvents(req.Params)
		resp := getEventsResponse{Events: events}
		if err != nil {
			resp.Error = err.Error()
		}
		if encErr := enc.Encode(resp); encErr != nil {
			return 1
		}
		if err != nil {
			return 1
		}
		return 0

	case "--authenticate":
		ah, ok := h.(AuthHandler)
		if !ok {
			_ = enc.Encode(authenticateResponse{Error: "plugin does not support authentication"})
			return 1
		}
		var req authenticateRequest
		if err := json.NewDecoder(in).Decode(&req); err != nil {
			_ = enc.Encode(authenticateResponse{Error: "decode request: " + err.Error()})
			return 1
		}
		tokens, err := ah.Authenticate(req.Params, os.Stderr)
		resp := authenticateResponse{Tokens: tokens}
		if err != nil {
			resp.Error = err.Error()
		}
		if encErr := enc.Encode(resp); encErr != nil {
			return 1
		}
		if err != nil {
			return 1
		}
		return 0

	default:
		return 1
	}
}

// ServeSetter dispatches a single plugin invocation for a setter binary.
// args should be os.Args[1:], in should be os.Stdin, out should be os.Stdout.
// Returns 0 on success, 1 on error. Plugin binaries should os.Exit with this value.
func ServeSetter(h SetterHandler, args []string, in io.Reader, out io.Writer) int {
	if len(args) == 0 {
		return 1
	}
	enc := json.NewEncoder(out)
	switch args[0] {
	case "--metadata":
		m := h.Metadata()
		resp := metadataResponse{
			Name:         m.Name,
			Type:         "setter",
			ParamSpecs:   m.ParamSpecs,
			SupportsAuth: m.SupportsAuth,
		}
		if _, ok := h.(AuthHandler); ok {
			resp.SupportsAuth = true
		}
		if err := enc.Encode(resp); err != nil {
			return 1
		}
		return 0

	case "--set-status":
		var req setStatusRequest
		if err := json.NewDecoder(in).Decode(&req); err != nil {
			return 1
		}
		result, err := h.SetStatus(req.Params)
		resp := setStatusResponse{
			Skipped:  result.Skipped,
			Response: result.Response,
		}
		if err != nil {
			resp.Error = err.Error()
		}
		if encErr := enc.Encode(resp); encErr != nil {
			return 1
		}
		if err != nil {
			return 1
		}
		return 0

	case "--authenticate":
		ah, ok := h.(AuthHandler)
		if !ok {
			_ = enc.Encode(authenticateResponse{Error: "plugin does not support authentication"})
			return 1
		}
		var req authenticateRequest
		if err := json.NewDecoder(in).Decode(&req); err != nil {
			_ = enc.Encode(authenticateResponse{Error: "decode request: " + err.Error()})
			return 1
		}
		tokens, err := ah.Authenticate(req.Params, os.Stderr)
		resp := authenticateResponse{Tokens: tokens}
		if err != nil {
			resp.Error = err.Error()
		}
		if encErr := enc.Encode(resp); encErr != nil {
			return 1
		}
		if err != nil {
			return 1
		}
		return 0

	default:
		return 1
	}
}
