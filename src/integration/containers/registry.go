package containers

import "sync"

var registry struct {
	mu sync.Mutex

	mysql *MySQLContainer
}
