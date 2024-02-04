package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	pkgauth "github.com/faustuzas/occa/src/pkg/auth"
	pkgclock "github.com/faustuzas/occa/src/pkg/clock"
	pkgid "github.com/faustuzas/occa/src/pkg/id"
	pkgmemstore "github.com/faustuzas/occa/src/pkg/memstore"
	pkgslices "github.com/faustuzas/occa/src/pkg/slices"
)

const (
	activeUserTTL         = 30 * time.Second
	activeUsersCollection = "active_users"
)

type ActiveUser struct {
	ID       pkgid.ID  `json:"id"`
	Username string    `json:"username"`
	LastSeen time.Time `json:"lastSeen"`
}

func (u *ActiveUser) marshall() ([]byte, error) {
	return json.Marshal(u)
}

func (u *ActiveUser) unmarshall(data []byte) error {
	return json.Unmarshal(data, u)
}

type ActiveUsersTracker interface {
	// HeartBeat marks the authenticated user as active.
	HeartBeat(ctx context.Context) error

	// ActiveUsers returns all known active users in the system.
	ActiveUsers(ctx context.Context) ([]ActiveUser, error)
}

type tracker struct {
	store pkgmemstore.Store
	clock pkgclock.Clock
}

func NewActiveUsersTracker(store pkgmemstore.Store, clock pkgclock.Clock) (ActiveUsersTracker, error) {
	return &tracker{
		store: store,
		clock: clock,
	}, nil
}

func (r *tracker) HeartBeat(ctx context.Context) error {
	var (
		principal = pkgauth.PrincipalFromContext(ctx)
		usr       = ActiveUser{
			ID:       principal.ID,
			Username: principal.UserName,
			LastSeen: r.clock.Now(),
		}
	)

	bytes, err := usr.marshall()
	if err != nil {
		return fmt.Errorf("marshaling active user: %w", err)
	}

	if err = r.store.SetCollectionItemWithTTL(ctx, activeUsersCollection, usr.ID.String(), bytes, activeUserTTL); err != nil {
		return fmt.Errorf("storing active user: %w", err)
	}

	return nil
}

func (r *tracker) ActiveUsers(ctx context.Context) ([]ActiveUser, error) {
	users, err := r.store.ListCollection(ctx, activeUsersCollection)
	if err != nil {
		return nil, fmt.Errorf("fetching users: %w", err)
	}

	return pkgslices.MapE(users, func(b []byte) (ActiveUser, error) {
		var usr ActiveUser
		if err := usr.unmarshall(b); err != nil {
			return ActiveUser{}, fmt.Errorf("unmarshaling active user: %w", err)
		}
		return usr, nil
	})
}
