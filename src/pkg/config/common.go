package config

import (
	"go.uber.org/zap"

	pkgservice "github.com/faustuzas/occa/src/pkg/service"
)

type CommonConfiguration struct {
	Logger *pkgservice.ExternalService[*zap.Logger, LoggerConfiguration] `yaml:"logger"`
}
