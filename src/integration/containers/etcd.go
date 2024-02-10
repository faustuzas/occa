package containers

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type EtcdContainer struct {
	Container

	Username string
	Password string
	Port     int
}

func WithEtcd(t *testing.T) *EtcdContainer {
	c, err := resolve[*EtcdContainer]("etcd", t, func() (registeredContainer, error) {
		const port = "2379"

		c, err := testcontainers.GenericContainer(context.Background(), testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Name:  "occa-etcd-test",
				Image: "quay.io/coreos/etcd:v3.4.15",
				Cmd: []string{
					"etcd",
					"--advertise-client-urls", "http://0.0.0.0:2379",
					"--listen-client-urls", "http://0.0.0.0:2379",
				},
				ExposedPorts: []string{fmt.Sprintf("%s/tcp", port)},
				WaitingFor:   wait.ForLog("serving insecure client requests"),
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

		return &EtcdContainer{
			Container: Container{c: c},

			Username: "",
			Password: "",
			Port:     mappedPort.Int(),
		}, nil
	})
	require.NoError(t, err)
	return c
}
