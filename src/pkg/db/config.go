package db

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const (
	defaultDBType = "mysql"
	defaultHost   = "localhost"
	defaultPort   = 3306
)

type Configuration struct {
	DBType string `yaml:"dbType"`

	// DataSourceName is a connection string accepted by sql.Open call. If it is provided,
	// all other connection details are ignored.
	DataSourceName string `yaml:"dsn"`

	Host string `yaml:"host"`
	Port int    `yaml:"port"`

	Database string `yaml:"database"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

func (c Configuration) validateAndSetDefaults() error {
	if c.DBType == "" {
		c.DBType = defaultDBType
	}

	if c.Host == "" {
		c.Host = defaultHost
	}

	if c.Port == 0 {
		c.Port = defaultPort
	}

	return nil
}

func (c Configuration) createDialector() (gorm.Dialector, error) {
	dsn := c.DataSourceName

	switch c.DBType {
	case "mysql":
		if dsn == "" {
			dsn = MysqlDNS(c.Username, c.Password, c.Host, c.Port, c.Database)
		}
		return mysql.Open(dsn), nil
	default:
		return nil, fmt.Errorf("database type %s not supported", c.DBType)
	}
}

func (c Configuration) Build() (*gorm.DB, error) {
	if err := c.validateAndSetDefaults(); err != nil {
		return nil, fmt.Errorf("validating config: %w", err)
	}

	dialector, err := c.createDialector()
	if err != nil {
		return nil, fmt.Errorf("constructing dialector: %w", err)
	}

	return gorm.Open(dialector, &gorm.Config{})
}
