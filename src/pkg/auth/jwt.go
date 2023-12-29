package auth

import (
	"crypto/rsa"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	jwtIssuerDomain = "faustasbutkus.eu"
	tokenDuration   = 24 * time.Hour
)

var (
	signingMethod = jwt.SigningMethodRS256
)

var (
	_ TokenValidator = (*JWTValidator)(nil)
	_ TokenIssuer    = (*JWTIssuer)(nil)
)

type claims struct {
	jwt.RegisteredClaims
	Principal
}

func NewJWTValidator(key *rsa.PublicKey) *JWTValidator {
	return &JWTValidator{
		key:       key,
		validator: jwt.NewValidator(jwt.WithExpirationRequired()),
	}
}

type JWTValidator struct {
	key       *rsa.PublicKey
	validator *jwt.Validator
}

func (v *JWTValidator) Validate(token string) (Principal, error) {
	t, err := jwt.ParseWithClaims(token, &claims{}, func(_ *jwt.Token) (interface{}, error) {
		return v.key, nil
	})
	if err != nil {
		return Principal{}, err
	}

	c := t.Claims.(*claims)
	if err = v.validator.Validate(c); err != nil {
		return Principal{}, err
	}

	return c.Principal, nil
}

func NewJWTIssuer(key *rsa.PrivateKey, now func() time.Time) *JWTIssuer {
	return &JWTIssuer{
		key: key,
		now: now,
	}
}

type JWTIssuer struct {
	key *rsa.PrivateKey
	now func() time.Time
}

func (i *JWTIssuer) Issue(p Principal) (string, error) {
	t := jwt.NewWithClaims(signingMethod, claims{
		Principal: p,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    jwtIssuerDomain,
			ExpiresAt: jwt.NewNumericDate(i.now().Add(tokenDuration)),
		},
	})

	return t.SignedString(i.key)
}
