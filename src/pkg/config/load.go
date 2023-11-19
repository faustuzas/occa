package config

import (
	"fmt"
	"os"

	"go.uber.org/config"
)

func LoadConfig[T any](configFile string) (T, error) {
	var zero T

	provider, err := config.NewYAML(config.Expand(os.LookupEnv), config.File(configFile))
	if err != nil {
		return zero, fmt.Errorf("unable to create YAML provider from config file: %w", err)
	}

	var cfg T
	if err = provider.Get(config.Root).Populate(&cfg); err != nil {
		return zero, fmt.Errorf("unable to populate from configuration file: %w", err)
	}
	return cfg, nil
}
