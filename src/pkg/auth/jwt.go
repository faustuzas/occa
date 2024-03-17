package auth

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
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

func (v *JWTValidator) Validate(_ context.Context, token string) (Principal, error) {
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

func (i *JWTIssuer) Issue(_ context.Context, p Principal) (string, error) {
	t := jwt.NewWithClaims(signingMethod, claims{
		Principal: p,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    jwtIssuerDomain,
			ExpiresAt: jwt.NewNumericDate(i.now().Add(tokenDuration)),
		},
	})

	return t.SignedString(i.key)
}

func ReadPublicKey(path string) (*rsa.PublicKey, error) {
	keyBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading public key: %w", err)
	}

	block, _ := pem.Decode(keyBytes)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block containing public key")
	}

	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaPubKey, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("expected *rsa.PublicKey, got %T", pubKey)
	}

	return rsaPubKey, nil
}

func ReadPrivateKey(path string) (*rsa.PrivateKey, error) {
	keyBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading private key: %w", err)
	}

	block, _ := pem.Decode(keyBytes)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block containing private key")
	}

	privKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parsing private key: %w", err)
	}

	rsaPrivKey, ok := privKey.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("could not cast %T to *rsa.PrivateKey", privKey)
	}

	return rsaPrivKey, nil
}
