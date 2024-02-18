package io

import (
	"context"
	"time"

	multierr "github.com/hashicorp/go-multierror"
)

type Closer interface {
	Close(context context.Context) error
}

type CloseFn func(ctx context.Context) error

func (f CloseFn) Close(ctx context.Context) error {
	return f(ctx)
}

func CloseWithoutContext(f func() error) CloseFn {
	return func(_ context.Context) error {
		return f()
	}
}

type Closers []Closer

func (c Closers) Close(ctx context.Context) error {
	var mErr error
	for _, e := range c {
		if err := e.Close(ctx); err != nil {
			mErr = multierr.Append(mErr, err)
		}
	}
	return mErr
}

func (c Closers) CloseWithTimeout(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return c.Close(ctx)
}

type Starter interface {
	Start(context context.Context) error
}

type StartFn func(ctx context.Context) error

func (f StartFn) Start(ctx context.Context) error {
	return f(ctx)
}

type Starters []Starter

func (s Starters) Start(ctx context.Context) error {
	for _, e := range s {
		if err := e.Start(ctx); err != nil {
			return err
		}
	}
	return nil
}
