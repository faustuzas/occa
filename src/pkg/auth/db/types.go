package db

import (
	"context"

	pkgdb "github.com/faustuzas/occa/src/pkg/db"
	pkgio "github.com/faustuzas/occa/src/pkg/io"
)

//go:generate sh -c "mockgen -package=db -destination=db_mock.go . Users"

type User struct {
	pkgdb.BaseModel

	Username string `gorm:"unique;not null"`
	Password string `gorm:"not null"`
}

type Users interface {
	pkgio.Closer

	Create(ctx context.Context, u User) error
	FindByUsername(ctx context.Context, username string) (User, error)

	Start(ctx context.Context) error
}
