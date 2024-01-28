package gateway

import (
	"fmt"
	"io"

	multierr "github.com/hashicorp/go-multierror"

	pkgauth "github.com/faustuzas/occa/src/pkg/auth"
	httpmiddleware "github.com/faustuzas/occa/src/pkg/http/middleware"
	pkginmemorydb "github.com/faustuzas/occa/src/pkg/inmemorydb"
)

type Services struct {
	InMemoryDB         pkginmemorydb.Store
	HTTPAuthMiddleware httpmiddleware.Middleware
	AuthRegisterer     pkgauth.Registerer

	closers []io.Closer
}

func (p Params) StartServices() (_ Services, err error) {
	var closers []io.Closer
	defer func() {
		if err != nil {
			err = multierr.Append(err, Services{closers: closers}.Close())
		}
	}()
	var starters []func() error

	inMemoryDB, err := p.Configuration.InMemoryDB.GetService()
	if err != nil {
		return Services{}, fmt.Errorf("building redis client: %w", err)
	}
	closers = append(closers, inMemoryDB)

	var httpAuthMiddleware httpmiddleware.Middleware
	switch p.Configuration.Auth.Type {
	case pkgauth.ValidatorConfigurationNoop:
		httpAuthMiddleware = pkgauth.NoopMiddleware()
	case pkgauth.ValidatorConfigurationJWTRSA:
		validator, e := p.Configuration.Auth.JWTValidator.GetService()
		if e != nil {
			return Services{}, fmt.Errorf("building JWT RSA validator: %w", e)
		}

		httpAuthMiddleware = pkgauth.HTTPTokenAuthorizationMiddleware(p.Logger, validator)
	default:
		return Services{}, fmt.Errorf("auth not configured")
	}

	tokenIssuer, err := p.Registerer.TokenIssuer.GetService()
	if err != nil {
		return Services{}, fmt.Errorf("building JWT token issuer: %w", err)
	}

	usersDB, err := p.Registerer.UsersDB.GetService()
	if err != nil {
		return Services{}, fmt.Errorf("building auth db connection: %w", err)
	}
	starters = append(starters, usersDB.Start)
	closers = append(closers, usersDB)

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
		InMemoryDB:         inMemoryDB,
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
