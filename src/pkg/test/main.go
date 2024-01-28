package test

import (
	"os"

	"go.uber.org/goleak"
)

const (
	ciTestRunEnvKey = "CI_TEST"
)

func PackageMain(m interface{ Run() int }) {
	if _, ok := os.LookupEnv(ciTestRunEnvKey); ok {
		goleak.VerifyTestMain(m, goleak.Cleanup(func(exitCode int) {
			if exitCode != 0 {
				os.Exit(exitCode)
			}
		}))
		return
	}

	m.Run()
}
