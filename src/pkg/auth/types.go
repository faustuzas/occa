package auth

import (
	pkgid "github.com/faustuzas/occa/src/pkg/id"
)

//go:generate sh -c "mockgen -package=auth -destination=auth_mock.go . TokenValidator,TokenIssuer,Registerer"

type Principal struct {
	ID       pkgid.ID `json:"id"`
	UserName string   `json:"userName"`
}

type TokenValidator interface {
	Validate(token string) (Principal, error)
}

type TokenIssuer interface {
	Issue(principal Principal) (string, error)
}

type Registerer interface {
	Login(username, password string) (string, error)
	Register(username, password string) error
}
