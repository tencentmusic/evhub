package nsq

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc/status"

	"github.com/tencentmusic/evhub/pkg/errcode"
	"github.com/tencentmusic/evhub/pkg/log"
	"github.com/tencentmusic/evhub/pkg/nsq"
)

// Nsq is a type of connector
type Nsq struct {
	NsqClient nsq.Client
}

// New creates a new Nsq connector
func New(nodeList []string) (*Nsq, error) {
	nsqClient, err := nsq.New(nodeList)
	if err != nil {
		return nil, fmt.Errorf("new nsq client err:%v", err)
	}
	return &Nsq{
		NsqClient: nsqClient,
	}, nil
}

// SendDelayMsg is the stores that send delay events
func (s *Nsq) SendDelayMsg(ctx context.Context, topic string, key string, data []byte, delay time.Duration) error {
	if err := s.NsqClient.SyncSendMsg(topic, data, delay); err != nil {
		log.With(ctx).Errorf("topic=%s send data err=%s\n delay:%v", topic, err, delay)
		return status.Errorf(errcode.DelayProduceFail, "nsq producer failed")
	}
	log.With(ctx).Debugf("nsq send data topic:%v delay:%v succ!", topic, delay)
	return nil
}

// SendMsg is the stores that send real-time events
func (s *Nsq) SendMsg(ctx context.Context, topic string, key string, data []byte) error {
	if err := s.NsqClient.SyncSendMsg(topic, data, time.Duration(0)); err != nil {
		log.With(ctx).Errorf("topic=%s send data err=%s\n", topic, err)
		return status.Errorf(errcode.DelayProduceFail, "nsq producer failed")
	}
	log.With(ctx).Debugf("nsq send data topic:%v  succ!", topic)
	return nil
}

// Stop stops the nsq connector gracefully.
func (s *Nsq) Stop() error {
	s.NsqClient.StopProducerList()
	return nil
}
