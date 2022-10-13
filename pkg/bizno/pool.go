package bizno

import (
	"fmt"
	"math/rand"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// Pool is the interface for creating a new biz number generator pool
type Pool interface {
	NextID() (string, error)
}

// bizNoPool is the biz number generator pool
type bizNoPool struct {
	GeneratorList []Generator
}

// Config is the configuration for the biz number generator
type Config struct {
	// EtcdEndpoints is the list of endpoints to connect to the etcd
	EtcdEndpoints []string
	// EtcdDialTimeout is the name of the etcd dial timeout
	EtcdDialTimeout time.Duration
	// GeneratorNum is the number of generators
	GeneratorNum int
}

// NewPool creates a new biz number generator pool
func NewPool(c *Config) (Pool, error) {
	if c == nil {
		return nil, fmt.Errorf("config is nil")
	}
	if c.GeneratorNum == 0 {
		c.GeneratorNum = 1
	}
	// initialize etcd client
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   c.EtcdEndpoints,
		DialTimeout: c.EtcdDialTimeout,
	})
	if err != nil {
		return nil, err
	}

	p := &bizNoPool{
		GeneratorList: make([]Generator, 0, c.GeneratorNum),
	}
	// initialize generator pool
	for i := 0; i < c.GeneratorNum; i++ {
		generator, err := NewGenerator(cli)
		if err != nil {
			return nil, err
		}
		p.GeneratorList = append(p.GeneratorList, generator)
	}
	return p, nil
}

// NextID is used to return the next generation number
func (s *bizNoPool) NextID() (string, error) {
	rand.Seed(time.Now().UnixNano())
	i := 0
	if len(s.GeneratorList) > 1 {
		i = rand.Intn(len(s.GeneratorList))
	}

	return s.GeneratorList[i].NextID()
}
