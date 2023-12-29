package gateway

import (
	"fmt"

	pkgauth "github.com/faustuzas/occa/src/pkg/auth"
	httpmiddleware "github.com/faustuzas/occa/src/pkg/http/middleware"
	"github.com/faustuzas/occa/src/pkg/redis"
)

type Services struct {
	Redis              redis.Client
	HTTPAuthMiddleware httpmiddleware.Middleware
}

func (p Params) ConfigureServices() (Services, error) {
	redisClient, err := p.Configuration.Redis.GetService()
	if err != nil {
		return Services{}, fmt.Errorf("building redis client: %w", err)
	}

	var httpAuthMiddleware httpmiddleware.Middleware
	switch p.Configuration.Auth.Type {
	case pkgauth.ValidatorConfigurationNoop:
		httpAuthMiddleware = pkgauth.NoopMiddleware()
	case pkgauth.ValidatorConfigurationJWTRSA:
		validator, e := p.Configuration.Auth.BuildJWTRSAValidator()
		if e != nil {
			return Services{}, fmt.Errorf("building JWT RSA validator: %w", e)
		}

		httpAuthMiddleware = pkgauth.HTTPTokenAuthorizationMiddleware(p.Logger, validator)
	default:
		return Services{}, fmt.Errorf("auth not configured")
	}

	return Services{
		Redis:              redisClient,
		HTTPAuthMiddleware: httpAuthMiddleware,
	}, nil
}
