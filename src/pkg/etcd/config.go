package etcd

import (
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type Configuration struct {
	Username  string   `yaml:"username"`
	Password  string   `yaml:"password"`
	Endpoints []string `yaml:"endpoints"`
}

func (c Configuration) Build() (*clientv3.Client, error) {
	return clientv3.New(clientv3.Config{
		Username:    c.Username,
		Password:    c.Password,
		Endpoints:   c.Endpoints,
		DialTimeout: 5 * time.Second,
	})
}
