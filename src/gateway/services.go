package gateway

import (
	"fmt"
	"io"

	multierr "github.com/hashicorp/go-multierror"

	"github.com/faustuzas/occa/src/gateway/services"
	pkgauth "github.com/faustuzas/occa/src/pkg/auth"
	pkgclock "github.com/faustuzas/occa/src/pkg/clock"
	httpmiddleware "github.com/faustuzas/occa/src/pkg/http/middleware"
)

type Services struct {
	HTTPAuthMiddleware httpmiddleware.Middleware
	AuthRegisterer     pkgauth.Registerer
	ActiveUserTracker  services.ActiveUsersTracker

	closers []io.Closer
}

func (p Params) StartServices() (_ Services, err error) {
	var (
		starters []func() error
		closers  []io.Closer

		clock = pkgclock.RealClock{}
	)
	defer func() {
		if err != nil {
			err = multierr.Append(err, Services{closers: closers}.Close())
		}
	}()

	memStore, err := p.Configuration.MemStore.Build()
	if err != nil {
		return Services{}, fmt.Errorf("building redis client: %w", err)
	}
	closers = append(closers, memStore)

	var httpAuthMiddleware httpmiddleware.Middleware
	switch p.Configuration.Auth.Type {
	case pkgauth.ValidatorConfigurationNoop:
		httpAuthMiddleware = pkgauth.NoopMiddleware()
	case pkgauth.ValidatorConfigurationJWTRSA:
		validator, e := p.Configuration.Auth.JWTValidator.Build()
		if e != nil {
			return Services{}, fmt.Errorf("building JWT RSA validator: %w", e)
		}

		httpAuthMiddleware = pkgauth.HTTPTokenAuthorizationMiddleware(p.Logger, validator)
	default:
		return Services{}, fmt.Errorf("auth not configured")
	}

	tokenIssuer, err := p.Registerer.TokenIssuer.Build()
	if err != nil {
		return Services{}, fmt.Errorf("building JWT token issuer: %w", err)
	}

	usersDB, err := p.Registerer.Users.Build()
	if err != nil {
		return Services{}, fmt.Errorf("building auth db connection: %w", err)
	}
	starters = append(starters, usersDB.Start)
	closers = append(closers, usersDB)

	activeUsersTracker, err := services.NewActiveUsersTracker(memStore, clock)
	if err != nil {
		return Services{}, fmt.Errorf("building active users tracker: %w", err)
	}

	var sErr error
	for _, s := range starters {
		if e := s(); e != nil {
			sErr = multierr.Append(sErr, e)
		}
	}
	if sErr != nil {
		return Services{}, fmt.Errorf("starting services: %w", sErr)
	}

	return Services{
		ActiveUserTracker:  activeUsersTracker,
		HTTPAuthMiddleware: httpAuthMiddleware,
		AuthRegisterer:     pkgauth.NewRegisterer(usersDB, tokenIssuer),
		closers:            closers,
	}, nil
}

func (s Services) Close() error {
	var mErr error
	for _, c := range s.closers {
		if err := c.Close(); err != nil {
			mErr = multierr.Append(mErr, err)
		}
	}
	return mErr
}
