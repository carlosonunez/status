package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"gopkg.in/yaml.v3"
)

const appName = "status"

// ErrNotFound is returned by Load when the config file does not exist.
var ErrNotFound = errors.New("config file not found")

// DefaultPath returns the OS-appropriate default path for the Status config file.
// On Linux/macOS this is $XDG_CONFIG_HOME/status/config.yaml (typically
// ~/.config/status/config.yaml). On Windows it resolves via XDG as well.
func DefaultPath() string {
	return filepath.Join(xdg.ConfigHome, appName, "config.yaml")
}

// Load reads and parses the config file at path. If path is empty, DefaultPath
// is used. It returns a non-nil *Config on success.
func Load(path string) (*Config, error) {
	if path == "" {
		path = DefaultPath()
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("%w at %s: run 'status check config' for help", ErrNotFound, path)
		}
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	return &cfg, nil
}

// Validate checks a parsed Config for required fields and logical consistency.
// It returns a slice of human-readable error strings so callers can display all
// problems at once rather than one at a time.
func Validate(cfg *Config) []string {
	var errs []string

	for i, rule := range cfg.Rules {
		prefix := fmt.Sprintf("rules[%d]", i)

		if rule.Name == "" {
			errs = append(errs, fmt.Sprintf("%s: name is required", prefix))
		}
		if rule.EventGetter.Name == "" {
			errs = append(errs, fmt.Sprintf("%s: event_getter.name is required", prefix))
		}
		for j, t := range rule.Transforms {
			tp := fmt.Sprintf("%s.transforms[%d]", prefix, j)
			if t.Name == "" {
				errs = append(errs, fmt.Sprintf("%s: name is required", tp))
			}
			if t.Pattern == "" {
				errs = append(errs, fmt.Sprintf("%s: pattern is required", tp))
			}
			if len(t.StatusSetters) == 0 {
				errs = append(errs, fmt.Sprintf("%s: at least one status_setter is required", tp))
			}
			for k, ss := range t.StatusSetters {
				if ss.Name == "" {
					errs = append(errs, fmt.Sprintf("%s.status_setters[%d]: name is required", tp, k))
				}
			}
		}
	}

	switch cfg.TokenStore.Backend {
	case "", "file":
		// file backend has no required fields
	case "dynamo":
		if cfg.TokenStore.Dynamo.TableName == "" {
			errs = append(errs, "token_store.dynamo.table_name is required when backend is 'dynamo'")
		}
		if cfg.TokenStore.Dynamo.Region == "" {
			errs = append(errs, "token_store.dynamo.region is required when backend is 'dynamo'")
		}
	default:
		errs = append(errs, fmt.Sprintf("token_store.backend: unknown backend %q (must be 'file' or 'dynamo')", cfg.TokenStore.Backend))
	}

	return errs
}
