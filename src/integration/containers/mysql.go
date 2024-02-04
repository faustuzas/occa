package containers

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"

	mysqldriver "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	pkgdb "github.com/faustuzas/occa/src/pkg/db"
)

type mysqlLogger struct {
	enableLogging atomic.Bool
}

func (m *mysqlLogger) Print(v ...interface{}) {
	if m.enableLogging.Load() {
		log.Println(v...)
	}
}

type MySQLContainer struct {
	Container

	Username string
	Password string
	Port     int
}

// WithMysql creates a docker MySQL container if it does not yet exist and returns an interface to interact with it.
func WithMysql(t *testing.T) *MySQLContainer {
	c, err := resolve[*MySQLContainer]("mysql", t, func() (registeredContainer, error) {
		// mysql driver is very noise regarding logs when booting up and checking for health
		logger := &mysqlLogger{}
		defer logger.enableLogging.Store(true)

		require.NoError(t, mysqldriver.SetLogger(logger))

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
		if err != nil {
			return nil, err
		}

		mappedPort, err := resolveMappedPort(c, port)
		if err != nil {
			return nil, fmt.Errorf("resolving mapped port: %w", err)
		}

		mysql := &MySQLContainer{
			Container: Container{c: c},
			Username:  username,
			Password:  password,
			Port:      mappedPort.Int(),
		}
		if err = mysql.waitForHealthy(); err != nil {
			return nil, err
		}

		return mysql, nil
	})
	require.NoError(t, err)
	return c
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
	db, err := sql.Open("mysql", c.DataSourceName(""))
	if err != nil {
		return "", fmt.Errorf("opening database connection: %w", err)
	}

	defer func() {
		_ = db.Close()
	}()

	databaseName := fmt.Sprintf("%v_%d", prefix, rand.Int())

	if _, err = db.Exec("CREATE DATABASE " + databaseName); err != nil {
		return "", fmt.Errorf("creating database: %w", err)
	}

	return databaseName, nil
}

func (c *MySQLContainer) DataSourceName(database string) string {
	return pkgdb.MysqlDNS(c.Username, c.Password, "localhost", c.Port, database)
}
