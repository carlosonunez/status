package config

// Config is the top-level structure for ~/.config/status/config.yaml.
type Config struct {
	TokenStore TokenStoreConfig `yaml:"token_store"`
	Rules      []RuleConfig     `yaml:"rules"`
}

// TokenStoreConfig selects and configures the credential storage backend.
type TokenStoreConfig struct {
	// Backend is "file" (default) or "dynamo".
	Backend string           `yaml:"backend"`
	File    FileStoreConfig  `yaml:"file,omitempty"`
	Dynamo  DynamoStoreConfig `yaml:"dynamo,omitempty"`
}

// FileStoreConfig configures the file-based token store.
type FileStoreConfig struct {
	// Path to the JSON token file. Defaults to ~/.config/status/tokens.json.
	Path string `yaml:"path,omitempty"`
}

// DynamoStoreConfig configures the AWS DynamoDB token store.
type DynamoStoreConfig struct {
	TableName string `yaml:"table_name"`
	Region    string `yaml:"region"`
	// Endpoint is optional. Set to point at DynamoDB Local (e.g. http://localhost:8000).
	Endpoint string `yaml:"endpoint,omitempty"`
}

// RuleConfig describes a single rule connecting an event getter to status setters.
type RuleConfig struct {
	Name        string            `yaml:"name"`
	Priority    int               `yaml:"priority"`
	EventGetter EventGetterConfig `yaml:"event_getter"`
	Transforms  []TransformConfig `yaml:"transforms"`
}

// EventGetterConfig names an event getter and its parameters.
type EventGetterConfig struct {
	Name   string                 `yaml:"name"`
	Params map[string]interface{} `yaml:"params,omitempty"`
}

// TransformConfig describes one transform within a rule.
type TransformConfig struct {
	Name          string               `yaml:"name"`
	Pattern       string               `yaml:"pattern"`
	StatusSetters []StatusSetterConfig `yaml:"status_setters"`
}

// StatusSetterConfig names a status setter and its parameters.
type StatusSetterConfig struct {
	Name   string                 `yaml:"name"`
	Params map[string]interface{} `yaml:"params,omitempty"`
}
