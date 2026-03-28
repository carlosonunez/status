// This file is compiled only during `go test`. It exposes internal
// constructors to the plugin_test package so tests can inject fake invokers
// without touching the filesystem.
package plugin

import (
	"context"
	"io"
	"io/fs"

	"github.com/carlosonunez/status/internal/getter"
	"github.com/carlosonunez/status/internal/pluginspec"
	"github.com/carlosonunez/status/internal/setter"
)

// InvokerFunc is the function signature expected by the test constructors below.
// It matches the unexported invoker type.
type InvokerFunc func(ctx context.Context, path, flag string, stdin io.Reader) ([]byte, error)

// NewExternalGetterWithInvoker creates an ExternalGetter using inv for both
// metadata fetching and GetEvents. No os.Stat check is performed, so no real
// binary is needed.
func NewExternalGetterWithInvoker(path string, inv InvokerFunc) (*ExternalGetter, error) {
	meta, err := fetchMetadata(path, invoker(inv))
	if err != nil {
		return nil, err
	}
	return newExternalGetterFromMeta(path, meta, invoker(inv))
}

// NewExternalSetterWithInvoker creates an ExternalSetter using inv for both
// metadata fetching and SetStatus. No os.Stat check is performed.
func NewExternalSetterWithInvoker(path string, inv InvokerFunc) (*ExternalSetter, error) {
	meta, err := fetchMetadata(path, invoker(inv))
	if err != nil {
		return nil, err
	}
	return newExternalSetterFromMeta(path, meta, invoker(inv))
}

// DiscoverFS exposes the internal discoverFS function for testing with
// an in-memory fs.FS (e.g. fstest.MapFS) and injected factory functions.
func DiscoverFS(
	fsys fs.FS,
	baseDir string,
	gr *getter.Registry,
	sr *setter.Registry,
	newGetter func(string) (*ExternalGetter, error),
	newSetter func(string) (*ExternalSetter, error),
) error {
	return discoverFS(fsys, baseDir, gr, sr, newGetter, newSetter)
}

// NewExternalGetterForTest builds an ExternalGetter directly from parts,
// bypassing metadata fetching. Useful when discovery tests need a factory
// that returns a pre-built getter without any invocation.
func NewExternalGetterForTest(name string, specs []pluginspec.ParamSpec, inv InvokerFunc) *ExternalGetter {
	return &ExternalGetter{name: name, paramSpecs: specs, invoke: invoker(inv)}
}

// NewExternalSetterForTest builds an ExternalSetter directly from parts.
func NewExternalSetterForTest(name string, specs []pluginspec.ParamSpec, inv InvokerFunc) *ExternalSetter {
	return &ExternalSetter{name: name, paramSpecs: specs, invoke: invoker(inv)}
}
