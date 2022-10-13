package pulsar

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/tencentmusic/evhub/pkg/redis"

	redigo "github.com/gomodule/redigo/redis"
	"github.com/tencentmusic/evhub/internal/processor/event"
	"github.com/tencentmusic/evhub/internal/processor/options"
	cc "github.com/tencentmusic/evhub/pkg/conf_connector"
	"github.com/tencentmusic/evhub/pkg/handler"
	"github.com/tencentmusic/evhub/pkg/log"
	"github.com/tencentmusic/evhub/pkg/util/routine"
	"github.com/tencentmusic/evhub/pkg/util/topicutil"

	"github.com/apache/pulsar-client-go/pulsar"
)

// Pulsar is a type of connector
type Pulsar struct {
	opts                 *options.Options
	c                    *options.PulsarConfig
	handler              *handler.Handler
	redisPool            *redigo.Pool
	client               pulsar.Client
	delayConsumerMap     sync.Map
	processorConsumerMap sync.Map
}

type EventType int

const (
	EventTypeProcessor EventType = 1
	EventTypeDelay     EventType = 2
)

// New creates a Pulsar connector
func New(opts *options.Options, c *options.PulsarConfig, handler *handler.Handler) (*Pulsar, error) {
	var client pulsar.Client
	var err error
	if c.Token == "" {
		client, err = pulsar.NewClient(pulsar.ClientOptions{
			URL: c.URL,
		})
	} else {
		client, err = pulsar.NewClient(pulsar.ClientOptions{
			URL:            c.URL,
			Authentication: pulsar.NewAuthenticationToken(c.Token),
		})
	}

	if err != nil {
		return nil, err
	}
	return &Pulsar{
		client:    client,
		opts:      opts,
		handler:   handler,
		c:         c,
		redisPool: redis.New(&opts.RedisConfig),
	}, nil
}

// KeyConsumer assembly consumer key
func KeyConsumer(topic, subName string) string {
	return fmt.Sprintf("%v_%v", topic, subName)
}

// Consumer is a consumer of Pulsar
type Consumer struct {
	consumer pulsar.Consumer
	event    *event.Event
}

// StoreProcessorConsumeMap stores the consumer of the processor
func (s *Pulsar) storeProcessorConsumeMap(dispatcherID string, consumer *Consumer) {
	s.processorConsumerMap.Store(dispatcherID, consumer)
}

// deleteProcessorConsumeMap delete the consumer of the processor
func (s *Pulsar) deleteProcessorConsumeMap(dispatcherID string) {
	s.processorConsumerMap.Delete(dispatcherID)
}

// getProcessorConsumeMap get the consumer of the processor
func (s *Pulsar) getProcessorConsumeMap(dispatcherID string) (*Consumer, bool) {
	v, ok := s.processorConsumerMap.Load(dispatcherID)
	if !ok {
		return nil, ok
	}
	vv, ok := v.(*Consumer)
	if !ok {
		return nil, ok
	}
	return vv, true
}

// setDelayConsumer set the delay consumer
func (s *Pulsar) setDelayConsumer(topic, subName string, consumer *Consumer) {
	s.delayConsumerMap.Store(KeyConsumer(topic, subName), consumer)
}

// deleteDelayConsumer delete the delay consumer
func (s *Pulsar) deleteDelayConsumer(topic, subName string) {
	s.delayConsumerMap.Delete(KeyConsumer(topic, subName))
}

// getDelayConsumer get the delay consumer
func (s *Pulsar) getDelayConsumer(topic, subName string) (*Consumer, bool) {
	v, ok := s.delayConsumerMap.Load(KeyConsumer(topic, subName))
	if !ok {
		return nil, ok
	}
	vv, ok := v.(*Consumer)
	if !ok {
		return nil, ok
	}
	return vv, true
}

// ConsumerProcessorTopic is the topic of the subscription processor
func (s *Pulsar) ConsumerProcessorTopic(confInfoList []*cc.ProcessorConfInfo) {
	defer routine.Recover()
	for i := range confInfoList {
		e, err := event.New(s.opts, s.handler, s.redisPool, confInfoList[i])
		if err != nil {
			log.Errorf("make event err:%v", err)
			return
		}
		s.subscribeProcessor(e, confInfoList[i])
	}
}

func (s *Pulsar) subscribeProcessor(e *event.Event, confInfo *cc.ProcessorConfInfo) {
	c := &Consumer{
		event: e,
	}
	s.storeProcessorConsumeMap(confInfo.DispatcherID, c)
	go s.subscribe(topicutil.Topic(confInfo.AppID, confInfo.TopicID),
		topicutil.SubName(confInfo.DispatcherID), pulsar.Shared, c)
	log.Infof("subscribe conf info:%+v", confInfo)
}

// WatchProcessorTopic watch the topic of the processor
func (s *Pulsar) WatchProcessorTopic(c *cc.ProcessorConfInfo, eventType int32) error {
	if eventType == cc.EventTypeDelete {
		if !s.isRunningProcessorConsumer(c.DispatcherID) {
			log.Infof("not running consumer conf:%+v", c)
			return nil
		}
		s.stopRunningProcessorConsumer(c.DispatcherID)
		s.deleteProcessorConsumeMap(c.DispatcherID)
		log.Infof("stop running consumer success info:%+v", c)
		return nil
	}
	if s.isRunningProcessorConsumer(c.DispatcherID) {
		s.stopRunningProcessorConsumer(c.DispatcherID)
		s.deleteProcessorConsumeMap(c.DispatcherID)
	}
	e, err := event.New(s.opts, s.handler, s.redisPool, c)
	if err != nil {
		log.Errorf("make event err:%v", err)
		return err
	}
	s.subscribeProcessor(e, c)
	return nil
}

// isRunningProcessorConsumer returns true if the processor is running
func (s *Pulsar) isRunningProcessorConsumer(dispatcherID string) bool {
	_, ok := s.getProcessorConsumeMap(dispatcherID)
	return ok
}

// stopRunningProcessorConsumer returns an error representing a stop error.  If is OK, returns nil.
func (s *Pulsar) stopRunningProcessorConsumer(dispatcherID string) {
	c, _ := s.getProcessorConsumeMap(dispatcherID)
	if c != nil && c.consumer != nil {
		c.consumer.Close()
		log.Infof("close consumer success")
		return
	}
}

// ConsumerDelayTopic is the subject of the subscription latency processor
func (s *Pulsar) ConsumerDelayTopic(confInfoList []*cc.ProducerConfInfo) {
	defer routine.Recover()
	for i := range confInfoList {
		e, err := event.New(s.opts, s.handler, s.redisPool, &cc.ProcessorConfInfo{})
		if err != nil {
			log.Errorf("make event err:%v", err)
			continue
		}
		s.subscribeDelay(e, confInfoList[i])
	}
}

func (s *Pulsar) subscribeDelay(e *event.Event, confInfo *cc.ProducerConfInfo) {
	c := &Consumer{
		event: e,
	}
	topic := topicutil.DelayTopic(confInfo.AppID, confInfo.TopicID)
	subName := topicutil.DelaySubName(confInfo.AppID, confInfo.TopicID)
	s.setDelayConsumer(topic, subName, c)
	go s.subscribe(topic,
		subName, pulsar.Shared, c)
	log.Infof("subscribe confInfo:%+v", confInfo)
}

// WatchDelayTopic watch the topic of the latency processor
func (s *Pulsar) WatchDelayTopic(c *cc.ProducerConfInfo, eventType int32) error {
	topic := topicutil.DelayTopic(c.AppID, c.TopicID)
	subName := topicutil.DelaySubName(c.AppID, c.AppID)

	if eventType == cc.EventTypeDelete {
		if !s.isRunningDelayConsumer(topic, subName) {
			log.Infof("stop success consumer topic:%v subName:%v", c.TopicID)
			return nil
		}
		s.stopDelayConsumer(topic, subName)
		return nil
	}
	if s.isRunningDelayConsumer(topic, subName) {
		s.stopDelayConsumer(topic, subName)
	}
	e, err := event.New(s.opts, s.handler, s.redisPool, &cc.ProcessorConfInfo{})
	if err != nil {
		log.Errorf("make event err:%v", err)
		return err
	}
	s.subscribeDelay(e, c)
	return nil
}

// isRunningDelayConsumer returns true if the delay processor is running
func (s *Pulsar) isRunningDelayConsumer(topic, subName string) bool {
	_, ok := s.getDelayConsumer(topic, subName)
	return ok
}

// stopDelayConsumer returns an error representing a stop error.  If is OK, returns nil.
func (s *Pulsar) stopDelayConsumer(topic string, subName string) {
	c, _ := s.getDelayConsumer(topic, subName)
	if c != nil && c.consumer != nil {
		c.consumer.Close()
		log.Infof("close consumer success")
	}
	s.deleteDelayConsumer(topic, subName)
}

// subscribe to the topic
func (s *Pulsar) subscribe(topic, subName string, subType pulsar.SubscriptionType, c *Consumer) {
	defer routine.Recover()
	consumer, err := s.client.Subscribe(pulsar.ConsumerOptions{
		Topic:            fmt.Sprintf("%v%v", s.c.TopicPrefix, topic),
		SubscriptionName: subName,
		Type:             subType,
	})
	if err != nil {
		if err.Error() == "connection error" {
			go log.Panicf("subscribe topic:%v subName:%v err:%v", topic, subName, err)
		}
		log.Errorf("subscribe topic:%v subName:%v err:%v", topic, subName, err)
		return
	}
	c.consumer = consumer
	go func() {
		defer routine.Recover()
		ctx := context.Background()
		for {
			msg, err := consumer.Receive(ctx)
			if err != nil {
				log.Errorf("received message err:%v", err)
				break
			}
			log.Debugf("received message:%+v", msg)
			if err = c.ProcessMessage(topic, msg); err != nil {
				log.Infof("process message err:%v", err)
				consumer.Nack(msg)
			}
			consumer.Ack(msg)
		}
	}()
}

// ProcessMessage processing message
func (s *Consumer) ProcessMessage(topic string, msg pulsar.Message) error {
	if strings.HasPrefix(topic, topicutil.DelayTopicPrefix) {
		return s.event.HandleEventInternal(msg.Payload())
	}
	return s.event.HandleEventExternal(msg.Payload())
}

// Stop stops the Pulsar connector gracefully.
func (s *Pulsar) Stop() error {
	s.delayConsumerMap.Range(func(key, value interface{}) bool {
		vv, ok := value.(*Consumer)
		if !ok {
			log.Errorf("convert err:%v", value)
			return true
		}
		if vv.consumer != nil {
			vv.consumer.Close()
		}
		log.Infof("stop consumer:%+v success ", vv)
		return true
	})
	if s.client != nil {
		s.client.Close()
	}
	log.Infof("stop pulsar consumer success")
	return nil
}
