package containers

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	pkgdb "github.com/faustuzas/occa/src/pkg/db"
)

type MySQLContainer struct {
	Container

	Username string
	Password string
	Port     int
}

// WithMysql creates a docker MySQL container if it does not yet exist and returns an interface to interact with it.
func WithMysql(t *testing.T) (mysql *MySQLContainer) {
	defer func() {
		if mysql != nil {
			t.Cleanup(func() {
				err := mysql.Terminate(context.Background())
				if err != nil {
					t.Errorf("mysql failed to terminate: %v", err)
				}
			})
		}
	}()

	registry.mu.Lock()
	defer registry.mu.Unlock()

	if m := registry.mysql; m != nil {
		registry.mysql.refCount++
		return registry.mysql
	}

	const (
		username = "root"
		password = "root"
		port     = "3306"
	)

	c, err := testcontainers.GenericContainer(context.Background(), testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Name:         "occa-mysql-test",
			Image:        "mysql:8.2",
			ExposedPorts: []string{fmt.Sprintf("%s/tcp", port)},
			WaitingFor:   wait.ForLog("ready for connections"),
			Env: map[string]string{
				"MYSQL_ROOT_PASSWORD": password,
			},
		},
		Started: true,
		Reuse:   true,
	})
	require.NoError(t, err)

	natPort, err := nat.NewPort("tcp", port)
	require.NoError(t, err)

	mappedPort, err := c.MappedPort(context.Background(), natPort)
	require.NoError(t, err)

	registry.mysql = &MySQLContainer{
		Container: Container{c: c, refCount: 1},
		Username:  username,
		Password:  password,
		Port:      mappedPort.Int(),
	}
	require.NoError(t, registry.mysql.waitForHealthy())

	fmt.Println("[test-container] mysql initialized and ready to use!")

	return registry.mysql
}

func (c *MySQLContainer) waitForHealthy() error {
	// sleep a little at the beginning to reduce the noise in the logs
	time.Sleep(1 * time.Second)

	const retryCount = 10

	var lastErr error
	for i := 0; i < retryCount; i++ {
		lastErr = func() error {
			db, err := sql.Open("mysql", pkgdb.MysqlDNS(c.Username, c.Password, "localhost", c.Port, ""))
			if err != nil {
				return fmt.Errorf("opening mysql connection: %w", err)
			}

			defer func() {
				_ = db.Close()
			}()

			if _, err = db.Exec("SELECT 1"); err != nil {
				return fmt.Errorf("executing health check: %w", err)
			}

			return nil
		}()

		if lastErr == nil {
			return nil
		}

		time.Sleep(500 * time.Millisecond)
	}

	return fmt.Errorf("timed out, last error: %w", lastErr)
}

func (c *MySQLContainer) WithTemporaryDatabase(cleanup interface{ Cleanup(func()) }, prefix string) (string, error) {
	db, err := sql.Open("mysql", pkgdb.MysqlDNS(c.Username, c.Password, "localhost", c.Port, ""))
	if err != nil {
		return "", fmt.Errorf("opening database connection: %w", err)
	}

	cleanup.Cleanup(func() {
		_ = db.Close()
	})

	databaseName := fmt.Sprintf("%v_%d", prefix, rand.Int())

	if _, err = db.Exec("CREATE DATABASE " + databaseName); err != nil {
		return "", fmt.Errorf("creating database: %w", err)
	}

	cleanup.Cleanup(func() {
		_, _ = db.Exec("DROP DATABASE " + databaseName)
	})

	return databaseName, nil
}
