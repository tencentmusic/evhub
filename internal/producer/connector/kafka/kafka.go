package kafka

import (
	"context"
	"time"

	"github.com/Shopify/sarama"

	"google.golang.org/grpc/status"

	"github.com/tencentmusic/evhub/pkg/errcode"
	"github.com/tencentmusic/evhub/pkg/log"
)

const (
	ProducerRetryMax = 3
	ProducerTimeout  = 10
)

// Kafka is a type of connector
type Kafka struct {
	// syncProducer is the name of the kafka producer
	syncProducer sarama.SyncProducer
	// brokerList is the name of the Kafka broker list
	brokerList []string
}

// New creates a new Kafka producer
func New(brokerList []string) (*Kafka, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = ProducerRetryMax
	config.Producer.Return.Successes = true
	config.Producer.Timeout = ProducerTimeout * time.Second
	config.Producer.Partitioner = sarama.NewHashPartitioner
	producer, err := sarama.NewSyncProducer(brokerList, config)
	if err != nil {
		return nil, err
	}

	return &Kafka{
		brokerList:   brokerList,
		syncProducer: producer,
	}, nil
}

// SendMsg is the stores that send real-time events
func (s *Kafka) SendMsg(ctx context.Context, topic string, orderingKey string, data []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(data),
		Key:   sarama.StringEncoder(orderingKey),
	}

	part, offset, err := s.syncProducer.SendMessage(msg)
	if err != nil {
		log.With(ctx).Errorf("topic=%s send message(%s) err=%s \n", topic, string(data), err)
		return status.Errorf(errcode.ProduceFail, "kafka producer failed")
	}
	log.With(ctx).Debugf("send data succ，partition=%d, offset=%d \n", part, offset)
	return nil
}

// SendDelayMsg Kafka
// Kafka does not support delayed messages
func (s *Kafka) SendDelayMsg(ctx context.Context, topic string, orderingKey string, data []byte, delay time.Duration) error {
	return nil
}

// Stop stops the kafka connector gracefully.
func (s *Kafka) Stop() error {
	if err := s.syncProducer.Close(); err != nil {
		return err
	}
	return nil
}
