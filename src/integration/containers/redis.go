package containers

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type RedisContainer struct {
	Container

	Port int
}

func WithRedis(t *testing.T) *RedisContainer {
	c, err := resolve[*RedisContainer]("redis", t, func() (registeredContainer, error) {
		const port = "6379"

		c, err := testcontainers.GenericContainer(context.Background(), testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Name:         "occa-redis-test",
				Image:        "redis:7.2.3",
				ExposedPorts: []string{fmt.Sprintf("%s/tcp", port)},
				WaitingFor:   wait.ForLog("Ready to accept connections tcp"),
			},
			Started: true,
			Reuse:   true,
		})
		if err != nil {
			return nil, err
		}

		mappedPort, err := resolveMappedPort(c, port)
		if err != nil {
			return nil, fmt.Errorf("resolving mapped port: %w", err)
		}

		return &RedisContainer{
			Container: Container{c: c},
			Port:      mappedPort.Int(),
		}, nil
	})
	require.NoError(t, err)
	return c
}
