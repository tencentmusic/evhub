package routine_pool

import (
	"sync"

	"github.com/panjf2000/ants/v2"
)

// Pool is a goroutine pool t
type Pool struct {
	PoolWithFunc *ants.PoolWithFunc
	wg           sync.WaitGroup
}

// New creates a new goroutine pool
func New(size int, pf func(interface{}), options ...ants.Option) (*Pool, error) {
	s := &Pool{}
	p, err := ants.NewPoolWithFunc(size, func(args interface{}) {
		defer s.wg.Done()
		pf(args)
	}, options...)
	if err != nil {
		return nil, err
	}
	s.PoolWithFunc = p
	return s, err
}

// Tune changes the capacity of this pool, note that it is noneffective to the infinite or pre-allocation pool.
func (s *Pool) Tune(size int) {
	s.PoolWithFunc.Tune(size)
}

// Invoke submits a task to pool.
//
// Note that you are allowed to call Pool.Invoke() from the current Pool.Invoke(),
// but what calls for special attention is that you will get blocked with the latest
// Pool.Invoke() call once the current Pool runs out of its capacity, and to avoid this,
// you should instantiate a PoolWithFunc with ants.WithNonblocking(true).
func (s *Pool) Invoke(args interface{}) error {
	s.wg.Add(1)
	return s.PoolWithFunc.Invoke(args)
}

// Close closes this pool and releases the worker queue.
func (s *Pool) Close() error {
	s.wg.Wait()
	s.PoolWithFunc.Release()
	return nil
}
