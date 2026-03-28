package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/carlosonunez/status/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte(content), 0600))
	return path
}

func TestLoad_ValidMinimalConfig(t *testing.T) {
	path := writeTempConfig(t, `
rules:
  - name: Test rule
    event_getter:
      name: google_calendar
    transforms:
      - name: In a meeting
        pattern: ".*"
        status_setters:
          - name: slack
            params:
              status_message: In a meeting
              emoji: "📆"
`)
	cfg, err := config.Load(path)
	require.NoError(t, err)
	require.Len(t, cfg.Rules, 1)
	assert.Equal(t, "Test rule", cfg.Rules[0].Name)
	assert.Equal(t, "google_calendar", cfg.Rules[0].EventGetter.Name)
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := config.Load("/nonexistent/path/config.yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "config file not found")
}

func TestLoad_InvalidYAML(t *testing.T) {
	path := writeTempConfig(t, `this: is: not: valid: yaml:`)
	_, err := config.Load(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parsing config file")
}

func TestLoad_EmptyPath_UsesDefault(t *testing.T) {
	// DefaultPath should never be empty
	assert.NotEmpty(t, config.DefaultPath())
}

func TestLoad_TokenStoreDefaults(t *testing.T) {
	path := writeTempConfig(t, `rules: []`)
	cfg, err := config.Load(path)
	require.NoError(t, err)
	// backend should be empty (caller treats as "file")
	assert.Empty(t, cfg.TokenStore.Backend)
}

func TestValidate_ValidConfig(t *testing.T) {
	cfg := &config.Config{
		Rules: []config.RuleConfig{
			{
				Name:        "My rule",
				EventGetter: config.EventGetterConfig{Name: "google_calendar"},
				Transforms: []config.TransformConfig{
					{
						Name:    "catch-all",
						Pattern: ".*",
						StatusSetters: []config.StatusSetterConfig{
							{Name: "slack"},
						},
					},
				},
			},
		},
	}
	errs := config.Validate(cfg)
	assert.Empty(t, errs)
}

func TestValidate_MissingRuleName(t *testing.T) {
	cfg := &config.Config{
		Rules: []config.RuleConfig{
			{
				EventGetter: config.EventGetterConfig{Name: "google_calendar"},
				Transforms: []config.TransformConfig{
					{Name: "t", Pattern: ".*", StatusSetters: []config.StatusSetterConfig{{Name: "slack"}}},
				},
			},
		},
	}
	errs := config.Validate(cfg)
	require.Len(t, errs, 1)
	assert.Contains(t, errs[0], "name is required")
}

func TestValidate_MissingEventGetterName(t *testing.T) {
	cfg := &config.Config{
		Rules: []config.RuleConfig{
			{
				Name: "My rule",
				Transforms: []config.TransformConfig{
					{Name: "t", Pattern: ".*", StatusSetters: []config.StatusSetterConfig{{Name: "slack"}}},
				},
			},
		},
	}
	errs := config.Validate(cfg)
	require.Len(t, errs, 1)
	assert.Contains(t, errs[0], "event_getter.name is required")
}

func TestValidate_MissingTransformPattern(t *testing.T) {
	cfg := &config.Config{
		Rules: []config.RuleConfig{
			{
				Name:        "My rule",
				EventGetter: config.EventGetterConfig{Name: "google_calendar"},
				Transforms: []config.TransformConfig{
					{Name: "t", StatusSetters: []config.StatusSetterConfig{{Name: "slack"}}},
				},
			},
		},
	}
	errs := config.Validate(cfg)
	require.Len(t, errs, 1)
	assert.Contains(t, errs[0], "pattern is required")
}

func TestValidate_NoStatusSetters(t *testing.T) {
	cfg := &config.Config{
		Rules: []config.RuleConfig{
			{
				Name:        "My rule",
				EventGetter: config.EventGetterConfig{Name: "google_calendar"},
				Transforms:  []config.TransformConfig{{Name: "t", Pattern: ".*"}},
			},
		},
	}
	errs := config.Validate(cfg)
	require.Len(t, errs, 1)
	assert.Contains(t, errs[0], "at least one status_setter is required")
}

func TestValidate_UnknownTokenStoreBackend(t *testing.T) {
	cfg := &config.Config{
		TokenStore: config.TokenStoreConfig{Backend: "redis"},
	}
	errs := config.Validate(cfg)
	require.Len(t, errs, 1)
	assert.Contains(t, errs[0], "unknown backend")
}

func TestValidate_DynamoMissingTableName(t *testing.T) {
	cfg := &config.Config{
		TokenStore: config.TokenStoreConfig{
			Backend: "dynamo",
			Dynamo:  config.DynamoStoreConfig{Region: "us-east-1"},
		},
	}
	errs := config.Validate(cfg)
	require.Len(t, errs, 1)
	assert.Contains(t, errs[0], "table_name is required")
}

func TestValidate_DynamoMissingRegion(t *testing.T) {
	cfg := &config.Config{
		TokenStore: config.TokenStoreConfig{
			Backend: "dynamo",
			Dynamo:  config.DynamoStoreConfig{TableName: "status-tokens"},
		},
	}
	errs := config.Validate(cfg)
	require.Len(t, errs, 1)
	assert.Contains(t, errs[0], "region is required")
}

func TestValidate_MultipleErrors(t *testing.T) {
	cfg := &config.Config{
		Rules: []config.RuleConfig{
			{}, // missing name, getter, and transforms
		},
	}
	errs := config.Validate(cfg)
	assert.GreaterOrEqual(t, len(errs), 2)
}

func TestRoundTrip(t *testing.T) {
	original := `rules:
    - name: My rule
      priority: 1
      event_getter:
        name: google_calendar
        params:
            event_names: []
            allow_all_day_events: true
      transforms:
        - name: catch-all
          pattern: .*
          status_setters:
            - name: slack
              params:
                status_message: In a meeting
                emoji: 📆
token_store:
    backend: file
    file:
        path: /tmp/tokens.json
`
	path := writeTempConfig(t, original)
	cfg, err := config.Load(path)
	require.NoError(t, err)

	assert.Equal(t, "My rule", cfg.Rules[0].Name)
	assert.Equal(t, 1, cfg.Rules[0].Priority)
	assert.Equal(t, "google_calendar", cfg.Rules[0].EventGetter.Name)
	assert.Equal(t, "catch-all", cfg.Rules[0].Transforms[0].Name)
	assert.Equal(t, "slack", cfg.Rules[0].Transforms[0].StatusSetters[0].Name)
	assert.Equal(t, "file", cfg.TokenStore.Backend)
	assert.Equal(t, "/tmp/tokens.json", cfg.TokenStore.File.Path)
}
