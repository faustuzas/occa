package memstore

import (
	"context"
	"time"
)

//go:generate sh -c "mockgen -package=memstore -destination=memstore_mock.go . Store"

type Store interface {
	SetCollectionItemWithTTL(ctx context.Context, collection string, key string, value []byte, ttl time.Duration) error
	ListCollectionKeys(ctx context.Context, collection string) ([][]byte, error)

	Close() error
}
