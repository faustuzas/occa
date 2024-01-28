package auth

import (
	pkgid "github.com/faustuzas/occa/src/pkg/id"
)

var (
	noopPrincipal = Principal{
		ID:       pkgid.NewID(),
		UserName: "The_User",
	}
)

var _ TokenIssuer = noopAuth{}
var _ TokenValidator = noopAuth{}

type noopAuth struct {
}

func (a noopAuth) Issue(_ Principal) (string, error) {
	return "token", nil
}

func (a noopAuth) Validate(_ string) (Principal, error) {
	return noopPrincipal, nil
}
