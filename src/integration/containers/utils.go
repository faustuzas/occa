package containers

import (
	"context"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
)

func resolveMappedPort(container testcontainers.Container, port string) (nat.Port, error) {
	natPort, err := nat.NewPort("tcp", port)
	if err != nil {
		return "", err
	}

	mappedPort, err := container.MappedPort(context.Background(), natPort)
	if err != nil {
		return "", err
	}

	return mappedPort, nil
}
