package auth

import (
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"github.com/faustuzas/occa/src/pkg/auth/db"
	pkgerrors "github.com/faustuzas/occa/src/pkg/errors"
	pkgid "github.com/faustuzas/occa/src/pkg/id"
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

func (s *RegistererImpl) Login(ctx context.Context, username, password string) (string, error) {
	user, err := s.users.FindByUsername(ctx, username)
	if err != nil {
		return "", fmt.Errorf("fetching user: %w", err)
	}

	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
		return "", pkgerrors.ErrUnauthorized(fmt.Errorf("passwords do not match"))
	}

	token, err := s.tokenIssuer.Issue(ctx, Principal{
		ID:       pkgid.FromString(user.ID),
		UserName: username,
	})
	if err != nil {
		return "", fmt.Errorf("issuing token: %w", err)
	}

	return token, nil
}

func (s *RegistererImpl) Register(ctx context.Context, username, password string) error {
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hashing password: %w", err)
	}

	return s.users.Create(ctx, db.User{
		Username: username,
		Password: string(hashedPass),
	})
}

func (s *RegistererImpl) Close() error {
	return s.Close()
}
