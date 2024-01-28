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

type Container struct {
	c testcontainers.Container

	mu       sync.Mutex
	refCount int
}

func (c *Container) Terminate(ctx context.Context) error {
	c.mu.Lock()
	c.refCount--
	c.mu.Unlock()

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
