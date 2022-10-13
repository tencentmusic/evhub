package nsq

import (
	"math/rand"
	"time"

	"github.com/nsqio/go-nsq"
)

// Client is interface for nsq clusters
type Client interface {
	// SyncSendMsg is used to send a message to the server
	SyncSendMsg(topic string, data []byte, delay time.Duration) error
	// ConsumerMsg is used to consume messages from the server
	ConsumerMsg(topic, channel string, lookUpAddrs []string,
		handler nsq.Handler) (*nsq.Consumer, error)
	// StopProducerList is used to stop the producer list gracefully
	StopProducerList()
}

const (
	MaxInFlight = 9
	MaxAttempts = 65531
)

// ClientNsq is the client for nsq clusters
type ClientNsq struct {
	producerList []*nsq.Producer
	nodeList     []string
}

// New creates a new nsq client
func New(nodeList []string) (Client, error) {
	var producerList []*nsq.Producer
	config := nsq.NewConfig()
	for _, nodeInfo := range nodeList {
		producer, err := nsq.NewProducer(nodeInfo, config)
		if err != nil {
			return nil, err
		}
		producerList = append(producerList, producer)
	}
	return &ClientNsq{
		producerList: producerList,
		nodeList:     nodeList,
	}, nil
}

// ConsumerMsg is used to consume messages from the server
func (s *ClientNsq) ConsumerMsg(topic, channel string,
	lookUpAddrs []string, handler nsq.Handler) (*nsq.Consumer, error) {
	config := nsq.NewConfig()
	config.MaxInFlight = MaxInFlight
	config.MaxAttempts = MaxAttempts
	consumer, err := nsq.NewConsumer(topic, channel, config)
	if err != nil {
		return nil, err
	}
	consumer.AddHandler(handler)
	if err = consumer.ConnectToNSQLookupds(lookUpAddrs); err != nil {
		return nil, err
	}
	return consumer, nil
}

// SyncSendMsg is used to send a message to the server
func (s *ClientNsq) SyncSendMsg(topic string, data []byte, delay time.Duration) error {
	rand.Seed(time.Now().UnixNano())
	i := 0
	if len(s.producerList) > 1 {
		i = rand.Intn(len(s.producerList))
	}
	return s.producerList[i].DeferredPublish(topic, delay, data)
}

// StopProducerList is used to stop the producer list gracefully
func (s *ClientNsq) StopProducerList() {
	for _, producer := range s.producerList {
		producer.Stop()
	}
}
