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
type ExternalSetter struct {
	binaryPath string
	name       string
	paramSpecs []pluginspec.ParamSpec
}

// NewExternalSetter reads metadata from the binary at path and returns an
// ExternalSetter. Returns an error if the binary is missing or its metadata
// cannot be parsed.
func NewExternalSetter(path string) (*ExternalSetter, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("plugin: setter binary not found at %q: %w", path, err)
	}
	meta, err := fetchMetadata(path)
	if err != nil {
		return nil, fmt.Errorf("plugin: setter %q metadata: %w", path, err)
	}
	return &ExternalSetter{
		binaryPath: path,
		name:       meta.Name,
		paramSpecs: meta.ParamSpecs,
	}, nil
}

func (s *ExternalSetter) Name() string                    { return s.name }
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

	out, err := runBinary(ctx, s.binaryPath, "--set-status", &reqBuf)
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

// Ensure ExternalSetter satisfies the interface at compile time.
var _ setter.StatusSetter = (*ExternalSetter)(nil)
