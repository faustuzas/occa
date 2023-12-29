package auth

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

type ValidatorConfigurationType string

const (
	ValidatorConfigurationNoop   ValidatorConfigurationType = "noop"
	ValidatorConfigurationJWTRSA ValidatorConfigurationType = "jwt_rsa"
)

type ValidatorConfiguration struct {
	Type ValidatorConfigurationType `yaml:"type"`
	JWT  struct {
		PublicKeyPath string `yaml:"publicKeyPath"`
	} `yaml:"jwt"`
}

func (c ValidatorConfiguration) BuildJWTRSAValidator() (TokenValidator, error) {
	if c.Type != ValidatorConfigurationJWTRSA {
		return nil, fmt.Errorf("incorrect auth type %v", c.Type)
	}

	keyBytes, err := os.ReadFile(c.JWT.PublicKeyPath)
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
