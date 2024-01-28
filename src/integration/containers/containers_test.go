package containers

import (
	"testing"

	pkgtest "github.com/faustuzas/occa/src/pkg/test"
)

func TestMySQL(t *testing.T) {
	WithMysql(t)
}

func TestRedis(t *testing.T) {
	WithRedis(t)
}

func TestMain(m *testing.M) {
	pkgtest.PackageMain(m)
}
