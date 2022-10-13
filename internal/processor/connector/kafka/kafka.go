package kafka

import (
	"context"
	"sync"

	"github.com/tencentmusic/evhub/pkg/redis"

	"github.com/Shopify/sarama"
	redigo "github.com/gomodule/redigo/redis"
	"github.com/tencentmusic/evhub/internal/processor/event"
	"github.com/tencentmusic/evhub/internal/processor/options"
	cc "github.com/tencentmusic/evhub/pkg/conf_connector"
	"github.com/tencentmusic/evhub/pkg/handler"
	"github.com/tencentmusic/evhub/pkg/log"
	"github.com/tencentmusic/evhub/pkg/util/routine"
	"github.com/tencentmusic/evhub/pkg/util/topicutil"
)

// Kafka is a type of connector
type Kafka struct {
	processorConsumerMap sync.Map
	redisPool            *redigo.Pool
	handler              *handler.Handler
	config               *options.KafkaConfig
	opts                 *options.Options
}

// New creates a Kafka connector
func New(opts *options.Options, c *options.KafkaConfig, handler *handler.Handler) (*Kafka, error) {
	return &Kafka{
		config:    c,
		handler:   handler,
		opts:      opts,
		redisPool: redis.New(&opts.RedisConfig),
	}, nil
}

// StoreProcessorConsumeMap stores the consumer of the processor
func (s *Kafka) storeProcessorConsumeMap(dispatcherID string, consumer *ConsumerGroup) {
	s.processorConsumerMap.Store(dispatcherID, consumer)
}

// deleteProcessorConsumeMap delete the consumer of the processor
func (s *Kafka) deleteProcessorConsumeMap(dispatcherID string) {
	s.processorConsumerMap.Delete(dispatcherID)
}

// getProcessorConsumeMap get the consumer of the processor
func (s *Kafka) getProcessorConsumeMap(dispatcherID string) (*ConsumerGroup, bool) {
	v, ok := s.processorConsumerMap.Load(dispatcherID)
	if !ok {
		return nil, ok
	}
	vv, ok := v.(*ConsumerGroup)
	if !ok {
		return nil, ok
	}
	return vv, true
}

// ConsumerProcessorTopic is the topic of the subscription processor
func (s *Kafka) ConsumerProcessorTopic(confInfoList []*cc.ProcessorConfInfo) {
	defer routine.Recover()
	for i := range confInfoList {
		go s.subscribe(confInfoList[i])
		log.Infof("subscribe conf info:%+v", confInfoList[i])
	}
}

// makeConsumerGroup assembles the consumer group
func (s *Kafka) makeConsumerGroup(confInfo *cc.ProcessorConfInfo) (*ConsumerGroup, error) {
	e, err := event.New(s.opts, s.handler, s.redisPool, confInfo)
	if err != nil {
		return nil, err
	}
	return &ConsumerGroup{
		confInfo:                   confInfo,
		Event:                      e,
		RedisPool:                  s.redisPool,
		OffsetsAutoCommitEnable:    s.config.OffsetsAutoCommitEnable,
		FlowControlBatchNum:        s.config.FlowControlBatchNum,
		InitOffsetGetOffsetTimeout: s.config.InitOffsetGetOffsetTimeout,
		InitOffsetGetOffsetPeroid:  s.config.InitOffsetGetOffsetPeroid,
	}, nil
}

// subscribe to the topic
func (s *Kafka) subscribe(confInfo *cc.ProcessorConfInfo) {
	c, err := s.makeConsumerGroup(confInfo)
	if err != nil {
		log.Error("make consumer group err:%v", err)
		return
	}
	topic, groupID := topicutil.Topic(confInfo.AppID, confInfo.TopicID), topicutil.SubName(confInfo.DispatcherID)
	config := sarama.NewConfig()
	config.Version, config.Consumer.Return.Errors = kafkaVersions[s.config.Version], true
	consumerGroup, err := sarama.NewConsumerGroup(s.config.BrokerList, groupID, config)
	if err != nil {
		log.Errorf("failed to create consumer group err: %v", err)
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	c.cancelFunc, c.ConsumerGroup = cancel, consumerGroup
	routine.GoWithRecovery(func() {
		for err := range consumerGroup.Errors() {
			log.Errorf("consumer group topic:%v group:%v err:%v", topic, groupID, err)
		}
	})
	routine.GoWithRecovery(func() {
		for {
			if err := consumerGroup.Consume(ctx, []string{topic}, c); err != nil {
				if err == sarama.ErrClosedConsumerGroup {
					log.Infof("consumer topic:%v group:%v group err:%v", topic, groupID, err)
					break
				}
				log.Errorf("consumer  topic:%v group:%v group err:%v", topic, groupID, err)
			}
			log.Infof("client consumer:%v topic:%v group:%v", c, topic, groupID)
		}
	})
	s.storeProcessorConsumeMap(confInfo.DispatcherID, c)
	log.Infof("consumer up and running topic:%v group:%v", topic, groupID)
}

// ConsumerDelayTopic is the subject of the subscription latency processor
func (s *Kafka) ConsumerDelayTopic(confInfoList []*cc.ProducerConfInfo) {

}

// WatchDelayTopic watch the topic of the latency processor
func (s *Kafka) WatchDelayTopic(c *cc.ProducerConfInfo, eventType int32) error {
	return nil
}

// isRunningConsumer returns true if the processor is running
func (s *Kafka) isRunningConsumer(dispatcherID string) bool {
	_, ok := s.getProcessorConsumeMap(dispatcherID)
	return ok
}

// stopRunningConsumer returns an error representing a stop error.  If is OK, returns nil.
func (s *Kafka) stopRunningConsumer(dispatcherID string) error {
	c, _ := s.getProcessorConsumeMap(dispatcherID)
	if c != nil {
		if err := c.Stop(); err != nil {
			return err
		}
		log.Infof("close consumer success")
		return nil
	}
	return nil
}

// WatchProcessorTopic watch the topic of the processor
func (s *Kafka) WatchProcessorTopic(c *cc.ProcessorConfInfo, eventType int32) error {
	if eventType == cc.EventTypeDelete {
		if !s.isRunningConsumer(c.DispatcherID) {
			log.Infof("not running consumer conf:%+v", c)
			return nil
		}
		if err := s.stopRunningConsumer(c.DispatcherID); err != nil {
			log.Errorf("stop running consumer err:%v info:%+v", err, c)
		}
		s.deleteProcessorConsumeMap(c.DispatcherID)
		log.Infof("stop running consumer success info:%+v", c)
		return nil
	}
	if s.isRunningConsumer(c.DispatcherID) {
		if err := s.stopRunningConsumer(c.DispatcherID); err != nil {
			log.Errorf("stop running consumer err:%v info:%+v", err, c)
		}
	}
	s.subscribe(c)
	log.Infof("run consumer success info:%+v", c)
	return nil
}

// Stop stops the Kafka connector gracefully.
func (s *Kafka) Stop() error {
	s.processorConsumerMap.Range(func(key, value interface{}) bool {
		vv, ok := value.(*ConsumerGroup)
		if !ok {
			log.Errorf("convert err:%v", value)
			return true
		}
		if err := vv.Stop(); err != nil {
			log.Errorf("stop consumer err:%v", err)
			return true
		}
		log.Infof("stop consumer group success kafka consumer:%v", vv)
		return true
	})
	return nil
}
