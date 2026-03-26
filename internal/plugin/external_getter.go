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
type ExternalGetter struct {
	binaryPath  string
	name        string
	minInterval time.Duration
	paramSpecs  []pluginspec.ParamSpec
	invoke      invoker
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
		binaryPath:  path,
		name:        meta.Name,
		minInterval: interval,
		paramSpecs:  meta.ParamSpecs,
		invoke:      inv,
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

// Ensure ExternalGetter satisfies the interface at compile time.
var _ getter.EventGetter = (*ExternalGetter)(nil)
