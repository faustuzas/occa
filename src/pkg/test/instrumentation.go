package test

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	pkginstrument "github.com/faustuzas/occa/src/pkg/instrument"
)

var Instrumentation pkginstrument.Instrumentation

func init() {
	l, err := zap.NewDevelopmentConfig().Build()
	if err != nil {
		panic(fmt.Sprintf("failed to build logger: %v", err))
	}
	Instrumentation.Logger = l
	Instrumentation.Registerer = prometheus.NewRegistry()
}
