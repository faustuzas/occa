package inmemorydb

import (
	"context"
	"time"
)

//go:generate sh -c "mockgen -package=inmemorydb -destination=inmemorydb_mock.go . Store"

type Store interface {
	SetCollectionItemWithTTL(ctx context.Context, collection string, key string, value string, ttl time.Duration) error
	ListCollectionKeys(ctx context.Context, collection string) ([]string, error)

	Close() error
}
