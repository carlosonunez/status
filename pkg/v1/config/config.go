package config

import (
	"errors"
	"fmt"
	"strings"

	v1 "github.com/carlosonunez/status/api/v1"
	"gopkg.in/yaml.v2"
)

// Parse turns a YAML file into a v1.Config
func Parse(s string) (*v1.Config, error) {
	var cfg v1.Config
	if err := yaml.UnmarshalStrict([]byte(s), &cfg); err != nil {
		yErr, ok := err.(*yaml.TypeError)
		if !ok {
			return nil, fmt.Errorf("config file is invalid: %s", err)
		}
		return nil, fmt.Errorf("config file is invalid: %s",
			errors.New(strings.Join(yErr.Errors, ", and ")))
	}
	return &cfg, nil
}
