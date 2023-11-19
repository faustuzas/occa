package config

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type CommonConfiguration struct {
	Logger struct {
		Level           string `yaml:"level"`
		DevelopmentMode bool   `yaml:"developmentMode"`
	} `yaml:"logger"`
}

func (c CommonConfiguration) BuildLogger() (*zap.Logger, error) {
	config := zap.NewProductionConfig()
	if c.Logger.DevelopmentMode {
		config = zap.NewDevelopmentConfig()
	}

	level := "info"
	if l := c.Logger.Level; len(l) != 0 {
		level = l
	}

	zapLevel, err := zapcore.ParseLevel(level)
	if err != nil {
		return nil, fmt.Errorf("failed to parse logging level: %w", err)
	}
	config.Level = zap.NewAtomicLevelAt(zapLevel)

	logger, err := config.Build()
	if err != nil {
		return nil, fmt.Errorf("failed bulding logger: %v", err)
	}

	return logger, nil
}
