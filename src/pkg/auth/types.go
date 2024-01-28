package auth

import (
	"context"

	pkgid "github.com/faustuzas/occa/src/pkg/id"
)

//go:generate sh -c "mockgen -package=auth -destination=auth_mock.go . TokenValidator,TokenIssuer,Registerer"

type Principal struct {
	ID       pkgid.ID `json:"id"`
	UserName string   `json:"userName"`
}

type TokenValidator interface {
	Validate(ctx context.Context, token string) (Principal, error)
}

type TokenIssuer interface {
	Issue(ctx context.Context, principal Principal) (string, error)
}

type Registerer interface {
	Login(ctx context.Context, username, password string) (string, error)
	Register(ctx context.Context, username, password string) error
}
