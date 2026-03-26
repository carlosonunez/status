package plugin

import (
	"fmt"
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
// Non-plugin files and non-executable files are silently skipped.
func DiscoverAll(binDir string, gr *getter.Registry, sr *setter.Registry) error {
	entries, err := os.ReadDir(binDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("plugin: scan %q: %w", binDir, err)
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		path := filepath.Join(binDir, name)

		switch {
		case strings.HasPrefix(name, getterPrefix):
			g, err := NewExternalGetter(path)
			if err != nil {
				return fmt.Errorf("plugin: load getter %q: %w", name, err)
			}
			gr.Register(g)

		case strings.HasPrefix(name, setterPrefix):
			s, err := NewExternalSetter(path)
			if err != nil {
				return fmt.Errorf("plugin: load setter %q: %w", name, err)
			}
			sr.Register(s)
		}
	}
	return nil
}

// DiscoverAllDefault scans the default plugin directory ($XDG_CONFIG_HOME/status/bin)
// and registers discovered plugins in the package-level default registries.
func DiscoverAllDefault(binDir string) error {
	return DiscoverAll(binDir, getter.DefaultRegistry(), setter.DefaultRegistry())
}
