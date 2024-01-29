package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Configuration struct {
	Level           string `yaml:"level"`
	Component       string `yaml:"component"`
	DevelopmentMode bool   `yaml:"developmentMode"`
}

func (c Configuration) Build() (*zap.Logger, error) {
	config := zap.NewProductionConfig()
	if c.DevelopmentMode {
		config = zap.NewDevelopmentConfig()
	}

	level := "info"
	if l := c.Level; len(l) != 0 {
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

	if c.Component != "" {
		logger = logger.With(zap.String("component", c.Component))
	}

	return logger, nil
}
