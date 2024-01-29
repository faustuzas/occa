package config

import "github.com/faustuzas/occa/src/pkg/logger"

type CommonConfiguration struct {
	Logger logger.Configuration `yaml:"logger"`
}
