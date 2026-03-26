package plugin

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/carlosonunez/status/internal/getter"
	"github.com/carlosonunez/status/internal/setter"
)

const (
	getterPrefix = "status-getter-"
	setterPrefix = "status-setter-"
)

// DiscoverAll scans binDir for status-getter-* and status-setter-* executables,
// fetches their metadata, and registers them in the provided registries.
// A missing binDir is silently ignored.
func DiscoverAll(binDir string, gr *getter.Registry, sr *setter.Registry) error {
	return discoverFS(os.DirFS(binDir), binDir, gr, sr, NewExternalGetter, NewExternalSetter)
}

// DiscoverAllDefault scans the default plugin directory and registers
// discovered plugins in the package-level default registries.
func DiscoverAllDefault(binDir string) error {
	return DiscoverAll(binDir, getter.DefaultRegistry(), setter.DefaultRegistry())
}

// discoverFS is the testable core: it reads from fsys, constructs plugin
// instances via the injected factories, and registers them.
func discoverFS(
	fsys fs.FS,
	baseDir string,
	gr *getter.Registry,
	sr *setter.Registry,
	newGetter func(string) (*ExternalGetter, error),
	newSetter func(string) (*ExternalSetter, error),
) error {
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("plugin: scan %q: %w", baseDir, err)
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		path := filepath.Join(baseDir, name)

		switch {
		case strings.HasPrefix(name, getterPrefix):
			g, err := newGetter(path)
			if err != nil {
				return fmt.Errorf("plugin: load getter %q: %w", name, err)
			}
			gr.Register(g)

		case strings.HasPrefix(name, setterPrefix):
			s, err := newSetter(path)
			if err != nil {
				return fmt.Errorf("plugin: load setter %q: %w", name, err)
			}
			sr.Register(s)
		}
	}
	return nil
}
