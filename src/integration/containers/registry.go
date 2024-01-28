package containers

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/testcontainers/testcontainers-go"
)

const (
	cleanUpContainerEnvKey = "CLEANUP_TEST_CONTAINERS"
)

type registeredContainer interface {
	IncRef()
	Terminate(ctx context.Context) error
}

type registry struct {
	mu sync.Mutex

	containers map[string]registeredContainer
}

var r = registry{containers: map[string]registeredContainer{}}

func resolve[T registeredContainer](
	name string,
	cleanup interface{ Cleanup(f func()) },
	create func() (registeredContainer, error),
) (T, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.containers[name]; !ok {
		c, err := create()
		if err != nil {
			var zero T
			return zero, fmt.Errorf("creating container %s: %w", name, err)
		}

		r.containers[name] = c
	}

	c := r.containers[name]
	c.IncRef()

	cleanup.Cleanup(func() {
		_ = c.Terminate(context.Background())
	})

	return c.(T), nil
}

type Container struct {
	c testcontainers.Container

	mu       sync.Mutex
	refCount int
}

func (c *Container) IncRef() {
	c.mu.Lock()
	c.refCount++
	c.mu.Unlock()
}

func (c *Container) Terminate(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.refCount--
	if c.refCount == 0 {
		if _, ok := os.LookupEnv(cleanUpContainerEnvKey); !ok {
			return nil
		}

		name, err := c.c.Name(context.Background())
		if err != nil {
			name = "<failed to get name>"
		}

		fmt.Printf("shutting down %v container\n", name)
		return c.c.Terminate(ctx)
	}

	return nil
}
