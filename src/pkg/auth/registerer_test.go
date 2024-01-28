package auth

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/faustuzas/occa/src/pkg/auth/db"
	pkgdb "github.com/faustuzas/occa/src/pkg/db"
	pkgid "github.com/faustuzas/occa/src/pkg/id"
)

func TestRegistererRegister_HappyPath(t *testing.T) {
	var (
		ctrl = gomock.NewController(t)

		usersDB = db.NewMockUsers(ctrl)
	)

	var caughtUser db.User
	usersDB.EXPECT().Create(gomock.Any(), gomock.Any()).
		Do(func(_ context.Context, user db.User) {
			caughtUser = user
		})

	r := NewRegisterer(usersDB, nil)
	require.NoError(t, r.Register(context.Background(), "name", "password"))

	require.Equal(t, "name", caughtUser.Username)
	require.NotContains(t, caughtUser.Password, "password")

	require.NoError(t, bcrypt.CompareHashAndPassword([]byte(caughtUser.Password), []byte("password")))
}

func TestRegistererLogin_HappyPath(t *testing.T) {
	var (
		ctrl = gomock.NewController(t)

		usersDB     = db.NewMockUsers(ctrl)
		tokenIssuer = NewMockTokenIssuer(ctrl)

		userID = pkgid.NewID()
	)

	usersDB.EXPECT().FindByUsername(gomock.Any(), "name").
		Return(db.User{
			BaseModel: pkgdb.BaseModel{
				ID: userID.String(),
			},
			Username: "name",
			Password: "$2a$10$AvGIwrqmPgKpjfIchIfMq.YKjz/f3BAmCzG8Vz7t9KCfm6n8okQ6C",
		}, nil)

	tokenIssuer.EXPECT().Issue(gomock.Any(), Principal{
		ID:       userID,
		UserName: "name",
	}).Return("secret token", nil)

	r := NewRegisterer(usersDB, tokenIssuer)

	token, err := r.Login(context.Background(), "name", "password")
	require.NoError(t, err)
	require.Equal(t, "secret token", token)
}
