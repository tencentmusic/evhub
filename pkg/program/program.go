package program

import (
	"sync"
	"syscall"

	"github.com/judwhite/go-svc"
	"github.com/tencentmusic/evhub/pkg/log"
	"github.com/tencentmusic/evhub/pkg/util/routine"
)

type program struct {
	once sync.Once
	s    Svc
}

type Svc interface {
	Start() error
	Stop() error
}

// Run runs your Service.
//
// Run will block until one of the signals specified in sig is received or a provided context is done.
// If sig is empty syscall.SIGINT and syscall.SIGTERM are used by default.
func Run(s Svc) {
	prg := &program{
		s: s,
	}
	if err := svc.Run(prg, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM); err != nil {
		log.Panicf("program run failed: %v", err)
	}
}

// Init is called before the program/service is started
func (p *program) Init(env svc.Environment) error {
	return nil
}

// Start is called after Init
func (p *program) Start() error {
	go func() {
		defer routine.Recover()
		if err := p.s.Start(); err != nil {
			log.Panicf("run main failed: %v", err)
		}
	}()
	log.Infof("program started")
	return nil
}

// Stop is called in response to syscall.SIGINT, syscall.SIGTERM
func (p *program) Stop() error {
	p.once.Do(func() {
		if err := p.s.Stop(); err != nil {
			log.Errorf("producer stopped error: %v", err)
		}
	})
	return nil
}
