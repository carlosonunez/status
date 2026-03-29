package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/carlosonunez/status/internal/pluginspec"
	"github.com/carlosonunez/status/internal/setter"
)

// ExternalSetter implements setter.StatusSetter by invoking an external binary.
// When the binary reports supports_auth:true in its metadata it also
// implements pluginspec.Authenticator.
type ExternalSetter struct {
	binaryPath   string
	name         string
	paramSpecs   []pluginspec.ParamSpec
	supportsAuth bool
	invoke       invoker
}

// NewExternalSetter reads metadata from the binary at path and returns an
// ExternalSetter. Returns an error if the binary is missing or its metadata
// cannot be parsed.
func NewExternalSetter(path string) (*ExternalSetter, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("plugin: setter binary not found at %q: %w", path, err)
	}
	meta, err := fetchMetadata(path, execInvoker)
	if err != nil {
		return nil, fmt.Errorf("plugin: setter %q metadata: %w", path, err)
	}
	return newExternalSetterFromMeta(path, meta, execInvoker)
}

// newExternalSetterFromMeta builds an ExternalSetter from pre-fetched metadata.
func newExternalSetterFromMeta(path string, meta metadataResponse, inv invoker) (*ExternalSetter, error) {
	return &ExternalSetter{
		binaryPath:   path,
		name:         meta.Name,
		paramSpecs:   meta.ParamSpecs,
		supportsAuth: meta.SupportsAuth,
		invoke:       inv,
	}, nil
}

func (s *ExternalSetter) Name() string                       { return s.name }
func (s *ExternalSetter) ParamSpecs() []pluginspec.ParamSpec { return s.paramSpecs }

func (s *ExternalSetter) SetStatus(ctx context.Context, st setter.Status) (setter.SetResult, error) {
	reqMap := make(map[string]any, len(st.Params.Keys()))
	for _, k := range st.Params.Keys() {
		v, _ := st.Params.Value(k)
		reqMap[k] = v
	}
	var reqBuf bytes.Buffer
	if err := json.NewEncoder(&reqBuf).Encode(setStatusRequest{Params: reqMap}); err != nil {
		return setter.SetResult{}, fmt.Errorf("plugin: setter %q encode request: %w", s.name, err)
	}

	out, err := s.invoke(ctx, s.binaryPath, "--set-status", &reqBuf)
	if err != nil {
		return setter.SetResult{}, fmt.Errorf("plugin: setter %q invocation failed: %w", s.name, err)
	}

	var resp setStatusResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		return setter.SetResult{}, fmt.Errorf("plugin: setter %q bad response: %w", s.name, err)
	}
	if resp.Error != "" {
		return setter.SetResult{}, fmt.Errorf("plugin: setter %q: %s", s.name, resp.Error)
	}
	return setter.SetResult{Skipped: resp.Skipped, Response: resp.Response}, nil
}

// Authenticate implements pluginspec.Authenticator by invoking the binary with
// --authenticate. The binary runs the full auth flow (printing prompts to
// stderr so the user sees them) and returns acquired tokens as JSON.
// The caller is responsible for storing the returned tokens in the TokenStore.
func (s *ExternalSetter) Authenticate(ctx context.Context, store pluginspec.TokenStore) error {
	if !s.supportsAuth {
		return fmt.Errorf("plugin: setter %q does not support authentication", s.name)
	}
	var reqBuf bytes.Buffer
	if err := json.NewEncoder(&reqBuf).Encode(authenticateRequest{Params: map[string]any{}}); err != nil {
		return fmt.Errorf("plugin: setter %q encode auth request: %w", s.name, err)
	}

	out, err := s.invoke(ctx, s.binaryPath, "--authenticate", &reqBuf)
	if err != nil {
		return fmt.Errorf("plugin: setter %q authenticate invocation: %w", s.name, err)
	}

	var resp authenticateResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		return fmt.Errorf("plugin: setter %q bad authenticate response: %w", s.name, err)
	}
	if resp.Error != "" {
		return fmt.Errorf("plugin: setter %q authenticate: %s", s.name, resp.Error)
	}

	for k, v := range resp.Tokens {
		if err := store.Set(s.name, k, v); err != nil {
			return fmt.Errorf("plugin: setter %q store token %q: %w", s.name, k, err)
		}
	}
	return nil
}

// SupportsAuth reports whether this setter's binary implements the auth flow.
func (s *ExternalSetter) SupportsAuth() bool { return s.supportsAuth }

// Ensure ExternalSetter satisfies the interface at compile time.
var _ setter.StatusSetter = (*ExternalSetter)(nil)
