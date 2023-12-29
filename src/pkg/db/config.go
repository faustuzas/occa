package db

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Configuration struct {
	DBType string `yaml:"dbType"`

	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Database string `yaml:"database"`

	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

func (c Configuration) Build() (*gorm.DB, error) {
	var db *gorm.DB
	switch c.DBType {
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", c.Username, c.Password, c.Host, c.Port, c.Database)
		d, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			return nil, fmt.Errorf("connecting to mysql: %w", err)
		}
		db = d
	default:
		panic(fmt.Sprintf("database type %s not supported", c.DBType))
	}

	return db, nil
}
