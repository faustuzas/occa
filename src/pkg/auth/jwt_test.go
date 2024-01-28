package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	pkgid "github.com/faustuzas/occa/src/pkg/id"
)

func TestJWTRoundTrip(t *testing.T) {
	var (
		now                   = time.Now()
		privateKey, publicKey = generateRSAKeyPair(t, 4096)

		principal = Principal{
			ID:       pkgid.NewID(),
			UserName: "mr test",
		}
	)

	issuer := NewJWTIssuer(privateKey, func() time.Time {
		return now
	})

	token, err := issuer.Issue(context.Background(), principal)
	require.NoError(t, err)

	validator := NewJWTValidator(publicKey)
	resultPrincipal, err := validator.Validate(context.Background(), token)
	require.NoError(t, err)

	require.Equal(t, principal, resultPrincipal)
}

func generateRSAKeyPair(t *testing.T, size int) (*rsa.PrivateKey, *rsa.PublicKey) {
	privateKey, err := rsa.GenerateKey(rand.Reader, size)
	require.NoError(t, err)

	return privateKey, &privateKey.PublicKey
}
