package auth

import (
	"fmt"
	"time"

	"google.golang.org/grpc"

	"github.com/faustuzas/occa/src/pkg/auth/db"
	pkgdb "github.com/faustuzas/occa/src/pkg/db"
	httpmiddleware "github.com/faustuzas/occa/src/pkg/http/middleware"
	pkginstrument "github.com/faustuzas/occa/src/pkg/instrument"
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

func (c ValidatorConfiguration) BuildHTTPMiddleware(inst pkginstrument.Instrumentation) (httpmiddleware.Middleware, error) {
	switch c.Type {
	case ValidatorConfigurationNoop:
		return HTTPNoopMiddleware(), nil
	case ValidatorConfigurationJWTRSA:
		validator, err := c.JWTValidator.Build()
		if err != nil {
			return nil, fmt.Errorf("building JWT RSA validator: %w", err)
		}

		return HTTPTokenAuthorizationMiddleware(inst, validator), nil
	default:
		return nil, fmt.Errorf("auth not configured")
	}
}

func (c ValidatorConfiguration) BuildGRPCStreamInterceptor(inst pkginstrument.Instrumentation) (grpc.StreamServerInterceptor, error) {
	switch c.Type {
	case ValidatorConfigurationNoop:
		return GRPCStreamNoopInterceptor(), nil
	case ValidatorConfigurationJWTRSA:
		validator, err := c.JWTValidator.Build()
		if err != nil {
			return nil, fmt.Errorf("building JWT RSA validator: %w", err)
		}

		return GRPCStreamTokenAuthorizationInterceptor(inst, validator), nil
	default:
		return nil, fmt.Errorf("auth not configured")
	}
}

type JWTValidatorConfiguration struct {
	PublicKeyPath string `yaml:"publicKeyPath"`
}

func (c JWTValidatorConfiguration) Build() (*JWTValidator, error) {
	publicKey, err := ReadPublicKey(c.PublicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("reading public key: %w", err)
	}
	return NewJWTValidator(publicKey), nil
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
	privateKey, err := ReadPrivateKey(c.PrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("reading private key: %w", err)
	}
	return NewJWTIssuer(privateKey, time.Now), nil
}
