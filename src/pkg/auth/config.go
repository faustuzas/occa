package auth

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"time"

	"github.com/faustuzas/occa/src/pkg/auth/db"
	pkgdb "github.com/faustuzas/occa/src/pkg/db"
)

type ValidatorConfigurationType string

const (
	ValidatorConfigurationNoop   ValidatorConfigurationType = "noop"
	ValidatorConfigurationJWTRSA ValidatorConfigurationType = "jwtRSA"
)

type ValidatorConfiguration struct {
	Type         ValidatorConfigurationType `yaml:"type"`
	JWTValidator JWTValidatorConfiguration  `yaml:"jwt"`
}

type JWTValidatorConfiguration struct {
	PublicKeyPath string `yaml:"publicKeyPath"`
}

func (c JWTValidatorConfiguration) Build() (*JWTValidator, error) {
	keyBytes, err := os.ReadFile(c.PublicKeyPath)
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

	return NewJWTValidator(rsaPubKey), nil
}

type RegistererConfiguration struct {
	Users       UsersConfiguration       `yaml:"users"`
	TokenIssuer TokenIssuerConfiguration `yaml:"jwt"`
}

type UsersConfiguration struct {
	DB pkgdb.Configuration `yaml:"db"`
}

func (c UsersConfiguration) Build() (db.Users, error) {
	gormDB, err := c.DB.Build()
	if err != nil {
		return nil, err
	}

	return db.NewUsersDB(gormDB), nil
}

type TokenIssuerConfiguration struct {
	PrivateKeyPath string `yaml:"privateKeyPath"`
}

func (c TokenIssuerConfiguration) Build() (TokenIssuer, error) {
	keyBytes, err := os.ReadFile(c.PrivateKeyPath)
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

	return NewJWTIssuer(rsaPrivKey, time.Now), nil
}
