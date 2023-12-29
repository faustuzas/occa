package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestJWTRoundTrip(t *testing.T) {
	var (
		now                   = time.Now()
		privateKey, publicKey = generateRSAKeyPair(t, 4096)

		principal = Principal{
			ID:       100,
			UserName: "mr test",
		}
	)

	issuer := NewJWTIssuer(privateKey, func() time.Time {
		return now
	})

	token, err := issuer.Issue(principal)
	require.NoError(t, err)

	validator := NewJWTValidator(publicKey)
	resultPrincipal, err := validator.Validate(token)
	require.NoError(t, err)

	require.Equal(t, principal, resultPrincipal)
}

func generateRSAKeyPair(t *testing.T, size int) (*rsa.PrivateKey, *rsa.PublicKey) {
	privateKey, err := rsa.GenerateKey(rand.Reader, size)
	require.NoError(t, err)

	return privateKey, &privateKey.PublicKey
}
