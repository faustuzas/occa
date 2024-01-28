package db

import (
	pkgdb "github.com/faustuzas/occa/src/pkg/db"
)

//go:generate sh -c "mockgen -package=db -destination=db_mock.go . Users"

type User struct {
	pkgdb.BaseModel

	Username string `gorm:"unique;not null"`
	Password string `gorm:"not null"`
}

type Users interface {
	Create(u User) error
	FindByUsername(username string) (User, error)

	Start() error
	Close() error
}
