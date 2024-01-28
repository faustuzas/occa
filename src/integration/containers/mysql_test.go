package containers

import (
	"testing"

	pkgtest "github.com/faustuzas/occa/src/pkg/test"
)

func TestMySQLCleanLifecycle(t *testing.T) {
	WithMysql(t)
	WithMysql(t)
}

func TestMain(m *testing.M) {
	pkgtest.PackageMain(m)
}
