package db

import (
	pkgdb "github.com/faustuzas/occa/src/pkg/db"
)

//go:generate sh -c "mockgen -package=db -destination=db_mock.go . Users"

type User struct {
	pkgdb.BaseModel
}

type Users interface {
	Get() (User, error)

	Start() error
	Close() error
}
