package auth

import (
	"github.com/faustuzas/occa/src/pkg/auth/db"
)

var _ Registerer = (*RegistererImpl)(nil)

func NewRegisterer(users db.Users, tokenIssuer TokenIssuer) *RegistererImpl {
	return &RegistererImpl{
		users:       users,
		tokenIssuer: tokenIssuer,
	}
}

type RegistererImpl struct {
	users       db.Users
	tokenIssuer TokenIssuer
}

func (s *RegistererImpl) Login(username, password string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (s *RegistererImpl) Register(username, password string) error {
	//TODO implement me
	panic("implement me")
}

func (s *RegistererImpl) Close() error {
	return s.Close()
}
