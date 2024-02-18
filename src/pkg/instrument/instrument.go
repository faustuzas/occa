package instrument

import (
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

type Instrumentation struct {
	Logger     *zap.Logger
	Registerer prometheus.Registerer
}
