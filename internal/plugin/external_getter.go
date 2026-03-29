package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/carlosonunez/status/internal/getter"
	"github.com/carlosonunez/status/internal/params"
	"github.com/carlosonunez/status/internal/pluginspec"
)

// ExternalGetter implements getter.EventGetter by invoking an external binary.
// When the binary reports supports_auth:true in its metadata it also
// implements pluginspec.Authenticator.
type ExternalGetter struct {
	binaryPath   string
	name         string
	minInterval  time.Duration
	paramSpecs   []pluginspec.ParamSpec
	supportsAuth bool
	invoke       invoker
}

// NewExternalGetter reads metadata from the binary at path and returns an
// ExternalGetter. Returns an error if the binary is missing or its metadata
// cannot be parsed.
func NewExternalGetter(path string) (*ExternalGetter, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("plugin: getter binary not found at %q: %w", path, err)
	}
	meta, err := fetchMetadata(path, execInvoker)
	if err != nil {
		return nil, fmt.Errorf("plugin: getter %q metadata: %w", path, err)
	}
	return newExternalGetterFromMeta(path, meta, execInvoker)
}

// newExternalGetterFromMeta builds an ExternalGetter from pre-fetched metadata.
func newExternalGetterFromMeta(path string, meta metadataResponse, inv invoker) (*ExternalGetter, error) {
	interval := time.Duration(0)
	if meta.MinInterval != "" {
		var err error
		interval, err = time.ParseDuration(meta.MinInterval)
		if err != nil {
			return nil, fmt.Errorf("plugin: getter %q bad min_interval %q: %w", path, meta.MinInterval, err)
		}
	}
	return &ExternalGetter{
		binaryPath:   path,
		name:         meta.Name,
		minInterval:  interval,
		paramSpecs:   meta.ParamSpecs,
		supportsAuth: meta.SupportsAuth,
		invoke:       inv,
	}, nil
}

func (g *ExternalGetter) Name() string                       { return g.name }
func (g *ExternalGetter) MinInterval() time.Duration         { return g.minInterval }
func (g *ExternalGetter) ParamSpecs() []pluginspec.ParamSpec { return g.paramSpecs }

func (g *ExternalGetter) GetEvents(ctx context.Context, p params.Params) ([]getter.Event, error) {
	reqMap := make(map[string]any, len(p.Keys()))
	for _, k := range p.Keys() {
		v, _ := p.Value(k)
		reqMap[k] = v
	}
	var reqBuf bytes.Buffer
	if err := json.NewEncoder(&reqBuf).Encode(getEventsRequest{Params: reqMap}); err != nil {
		return nil, fmt.Errorf("plugin: getter %q encode request: %w", g.name, err)
	}

	out, err := g.invoke(ctx, g.binaryPath, "--get-events", &reqBuf)
	if err != nil {
		return nil, fmt.Errorf("plugin: getter %q invocation failed: %w", g.name, err)
	}

	var resp getEventsResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		return nil, fmt.Errorf("plugin: getter %q bad response: %w", g.name, err)
	}
	if resp.Error != "" {
		return nil, fmt.Errorf("plugin: getter %q: %s", g.name, resp.Error)
	}

	events := make([]getter.Event, 0, len(resp.Events))
	for _, e := range resp.Events {
		ge, err := convertEvent(e)
		if err != nil {
			return nil, fmt.Errorf("plugin: getter %q event decode: %w", g.name, err)
		}
		events = append(events, ge)
	}
	return events, nil
}

func convertEvent(e externalEvent) (getter.Event, error) {
	var starts, ends time.Time
	var err error
	if e.StartsAt != "" {
		starts, err = time.Parse(time.RFC3339, e.StartsAt)
		if err != nil {
			return getter.Event{}, fmt.Errorf("bad starts_at %q: %w", e.StartsAt, err)
		}
	}
	if e.EndsAt != "" {
		ends, err = time.Parse(time.RFC3339, e.EndsAt)
		if err != nil {
			return getter.Event{}, fmt.Errorf("bad ends_at %q: %w", e.EndsAt, err)
		}
	}
	return getter.Event{
		Calendar: e.Calendar,
		Title:    e.Title,
		StartsAt: starts,
		EndsAt:   ends,
		AllDay:   e.AllDay,
		Meta:     e.Meta,
	}, nil
}

// Authenticate implements pluginspec.Authenticator by invoking the binary with
// --authenticate. The binary runs the full auth flow (printing prompts to
// stderr so the user sees them) and returns acquired tokens as JSON.
// The caller is responsible for storing the returned tokens in the TokenStore.
func (g *ExternalGetter) Authenticate(ctx context.Context, store pluginspec.TokenStore) error {
	if !g.supportsAuth {
		return fmt.Errorf("plugin: getter %q does not support authentication", g.name)
	}
	var reqBuf bytes.Buffer
	if err := json.NewEncoder(&reqBuf).Encode(authenticateRequest{Params: map[string]any{}}); err != nil {
		return fmt.Errorf("plugin: getter %q encode auth request: %w", g.name, err)
	}

	out, err := g.invoke(ctx, g.binaryPath, "--authenticate", &reqBuf)
	if err != nil {
		return fmt.Errorf("plugin: getter %q authenticate invocation: %w", g.name, err)
	}

	var resp authenticateResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		return fmt.Errorf("plugin: getter %q bad authenticate response: %w", g.name, err)
	}
	if resp.Error != "" {
		return fmt.Errorf("plugin: getter %q authenticate: %s", g.name, resp.Error)
	}

	for k, v := range resp.Tokens {
		if err := store.Set(g.name, k, v); err != nil {
			return fmt.Errorf("plugin: getter %q store token %q: %w", g.name, k, err)
		}
	}
	return nil
}

// SupportsAuth reports whether this getter's binary implements the auth flow.
func (g *ExternalGetter) SupportsAuth() bool { return g.supportsAuth }

// Ensure ExternalGetter satisfies the interface at compile time.
var _ getter.EventGetter = (*ExternalGetter)(nil)
