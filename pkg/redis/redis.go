package redis

import (
	"time"

	redigo "github.com/gomodule/redigo/redis"
)

const (
	MaxIdle     = 1024
	IdleTimeout = 3 * time.Minute
)

// Config is the configuration for the redis clusters
type Config struct {
	// Addr is the name of redis address
	Addr string
	// Timeout is the name of the timeout for the cluster
	Timeout time.Duration
	// Password is the password for the cluster
	Password string
	// DB is the name of the database
	DB int
}

// New creates a new redis pool
func New(c *Config) *redigo.Pool {
	return &redigo.Pool{
		MaxIdle:     MaxIdle,
		IdleTimeout: IdleTimeout,
		Dial: func() (redigo.Conn, error) {
			return redigo.Dial("tcp",
				c.Addr,
				redigo.DialConnectTimeout(c.Timeout),
				redigo.DialWriteTimeout(c.Timeout),
				redigo.DialReadTimeout(c.Timeout),
				redigo.DialPassword(c.Password),
				redigo.DialDatabase(c.DB),
			)
		},
	}
}
