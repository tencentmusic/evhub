package pulsar

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/tencentmusic/evhub/pkg/log"
)

// Pulsar is a type of connector
type Pulsar struct {
	// client is the name of the pulsar client
	client pulsar.Client
	// producerMap is name of the producer map
	producerMap sync.Map
	// topicPrefix is the name of the topic prefix
	topicPrefix string
	// url is the name of the pulsar broker url
	url string
}

// New creates a new pulsar connector
func New(url, topicPrefix string, token string) (*Pulsar, error) {
	var client pulsar.Client
	var err error
	if token == "" {
		client, err = pulsar.NewClient(pulsar.ClientOptions{

			URL: url,
		})
	} else {
		client, err = pulsar.NewClient(pulsar.ClientOptions{
			URL:            url,
			Authentication: pulsar.NewAuthenticationToken(token),
		})
	}
	if err != nil {
		return nil, err
	}
	return &Pulsar{
		client:      client,
		url:         url,
		topicPrefix: topicPrefix,
	}, nil
}

// setProducer stores the producer of the pulsar
func (s *Pulsar) setProducer(topic string, p pulsar.Producer) {
	s.producerMap.Store(topic, p)
}

// deleteProducer delete the producer of the pulsar
func (s *Pulsar) deleteProducer(topic string) {
	s.producerMap.Delete(topic)
}

// getProducer get the producer of the pulsar
func (s *Pulsar) getProducer(topic string) (pulsar.Producer, error) {
	v, ok := s.producerMap.Load(topic)
	if !ok {
		producer, err := s.client.CreateProducer(pulsar.ProducerOptions{
			Topic:         fmt.Sprintf("%v%v", s.topicPrefix, topic),
			HashingScheme: pulsar.JavaStringHash,
		})
		if err != nil {
			return nil, err
		}
		s.setProducer(topic, producer)
		return producer, nil
	}
	vv, ok := v.(pulsar.Producer)
	if !ok {
		return nil, fmt.Errorf("convert err")
	}
	return vv, nil
}

// SendDelayMsg is the stores that send delay events
func (s *Pulsar) SendDelayMsg(ctx context.Context, topic string, orderingKey string, data []byte, delay time.Duration) error {
	p, err := s.getProducer(topic)
	if err != nil {
		return err
	}
	msgID, err := p.Send(ctx, &pulsar.ProducerMessage{
		Key:          orderingKey,
		Payload:      data,
		DeliverAfter: delay,
	})
	if err != nil {
		s.deleteProducer(topic)
		log.With(ctx).Errorf("topic=%s send data err=%s\n delay:%v msgID:%+v", topic, err, delay, msgID)
		return err
	}
	log.With(ctx).Debugf("pulsar send data succ topic=%s msgID:%+v delay:%v", topic, msgID, delay)
	return nil
}

// SendMsg is the stores that send real-time events
func (s *Pulsar) SendMsg(ctx context.Context, topic string, orderingKey string, data []byte) error {
	p, err := s.getProducer(topic)
	if err != nil {
		return err
	}
	msgID, err := p.Send(ctx, &pulsar.ProducerMessage{
		Key:     orderingKey,
		Payload: data,
	})
	if err != nil {
		log.With(ctx).Errorf("topic=%s send data err=%s\n msgID:%+v", topic, err, msgID)
		return err
	}
	log.With(ctx).Debugf("pulsar send data succ msgID:%+v", msgID)
	return nil
}

// Stop stops the pulsar connector gracefully.
func (s *Pulsar) Stop() error {
	s.producerMap.Range(func(key, value interface{}) bool {
		vv, ok := value.(pulsar.Producer)
		if !ok {
			log.Errorf("convert err:%v", value)
			return true
		}
		vv.Close()
		log.Infof("stop producer:%+v success ", vv)
		return true
	})
	if s.client != nil {
		s.client.Close()
	}
	return nil
}
