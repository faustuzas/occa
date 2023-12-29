package test

import (
	"fmt"

	"go.uber.org/zap"
)

var Logger *zap.Logger

func init() {
	l, err := zap.NewDevelopmentConfig().Build()
	if err != nil {
		panic(fmt.Sprintf("failed to build logger: %v", err))
	}
	Logger = l
}
