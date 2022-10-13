package nsq

import (
	"fmt"
	"strings"
	"sync"

	"github.com/tencentmusic/evhub/pkg/redis"

	redigo "github.com/gomodule/redigo/redis"
	go_nsq "github.com/nsqio/go-nsq"
	"github.com/tencentmusic/evhub/internal/processor/event"
	"github.com/tencentmusic/evhub/internal/processor/options"
	cc "github.com/tencentmusic/evhub/pkg/conf_connector"
	"github.com/tencentmusic/evhub/pkg/handler"
	"github.com/tencentmusic/evhub/pkg/log"
	"github.com/tencentmusic/evhub/pkg/nsq"
	"github.com/tencentmusic/evhub/pkg/util/routine"
	"github.com/tencentmusic/evhub/pkg/util/topicutil"
)

// Nsq is a type of connector
type Nsq struct {
	NsqClient            nsq.Client
	Opts                 *options.Options
	Handler              *handler.Handler
	RedisPool            *redigo.Pool
	delayConsumerMap     sync.Map
	processorConsumerMap sync.Map
}

// New creates a Nsq connector
func New(opts *options.Options, c *options.NsqConfig, handler *handler.Handler) (*Nsq, error) {
	nsqClient, err := nsq.New(c.NodeList)
	if err != nil {
		return nil, fmt.Errorf("new nsq client err:%v", err)
	}
	return &Nsq{
		NsqClient: nsqClient,
		Opts:      opts,
		Handler:   handler,
		RedisPool: redis.New(&opts.RedisConfig),
	}, nil
}

// KeyConsumer assembly consumer key
func KeyConsumer(topic, subName string) string {
	return fmt.Sprintf("%v_%v", topic, subName)
}

// setDelayConsumer set the delay consumer
func (s *Nsq) setDelayConsumer(topic, subName string, consumer *Handler) {
	s.delayConsumerMap.Store(KeyConsumer(topic, subName), consumer)
}

// deleteDelayConsumer delete the delay consumer
func (s *Nsq) deleteDelayConsumer(topic, subName string) {
	s.delayConsumerMap.Delete(KeyConsumer(topic, subName))
}

// getDelayConsumer get the delay consumer
func (s *Nsq) getDelayConsumer(topic, subName string) (*Handler, bool) {
	v, ok := s.delayConsumerMap.Load(KeyConsumer(topic, subName))
	if !ok {
		return nil, ok
	}
	vv, ok := v.(*Handler)
	if !ok {
		return nil, ok
	}
	return vv, true
}

// storeProcessorConsumeMap stores the consumer of the processor
func (s *Nsq) storeProcessorConsumeMap(dispatcherID string, consumer *Handler) {
	s.processorConsumerMap.Store(dispatcherID, consumer)
}

// deleteProcessorConsumeMap delete the consumer of the processor
func (s *Nsq) deleteProcessorConsumeMap(dispatcherID string) {
	s.processorConsumerMap.Delete(dispatcherID)
}

// getProcessorConsumeMap get the consumer of the processor
func (s *Nsq) getProcessorConsumeMap(dispatcherID string) (*Handler, bool) {
	v, ok := s.processorConsumerMap.Load(dispatcherID)
	if !ok {
		return nil, ok
	}
	vv, ok := v.(*Handler)
	if !ok {
		return nil, ok
	}
	return vv, true
}

// ConsumerDelayTopic is the subject of the subscription latency processor
func (s *Nsq) ConsumerDelayTopic(confInfoList []*cc.ProducerConfInfo) {
	defer routine.Recover()
	for i := range confInfoList {
		e, err := event.New(s.Opts, s.Handler, s.RedisPool, &cc.ProcessorConfInfo{})
		if err != nil {
			log.Errorf("make event err:%v", err)
			continue
		}
		s.subscribeDelay(e, confInfoList[i])
	}
}

func (s *Nsq) subscribeDelay(e *event.Event, confInfo *cc.ProducerConfInfo) {
	h := &Handler{}
	topic := topicutil.DelayTopic(confInfo.AppID, confInfo.TopicID)
	subName := topicutil.DelaySubName(confInfo.AppID, confInfo.TopicID)
	s.setDelayConsumer(topic, subName, h)
	go s.subscribe(topic, subName, e, h)
	log.Infof("subscribe confInfo:%+v", confInfo)
}

// subscribe to the topic
func (s *Nsq) subscribe(topic, subName string, e *event.Event, h *Handler) {
	h.Topic = topic
	h.Event = e
	h.Channel = subName
	consumer, err := s.NsqClient.ConsumerMsg(topic,
		subName, s.Opts.ConnectorConfig.NsqConfig.LookUpAddrs, h)
	if err != nil {
		log.Errorf("nsq consumer err:%v", err)
	}
	h.Consumer = consumer
	s.setDelayConsumer(topic, subName, h)
}

// Handler is a consumer of Nsq
type Handler struct {
	Topic    string
	Channel  string
	Event    *event.Event
	Consumer *go_nsq.Consumer
}

// HandleMessage processing message
func (s *Handler) HandleMessage(msg *go_nsq.Message) error {
	defer routine.Recover()
	if strings.HasPrefix(s.Topic, topicutil.DelayTopicPrefix) {
		return s.Event.HandleEventInternal(msg.Body)
	}
	return s.Event.HandleEventExternal(msg.Body)
}

// WatchDelayTopic watch the topic of the latency processor
func (s *Nsq) WatchDelayTopic(c *cc.ProducerConfInfo, eventType int32) error {
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
	e, err := event.New(s.Opts, s.Handler, s.RedisPool, &cc.ProcessorConfInfo{})
	if err != nil {
		log.Errorf("make event err:%v", err)
		return err
	}
	s.subscribeDelay(e, c)
	return nil
}

// isRunningDelayConsumer returns true if the delay processor is running
func (s *Nsq) isRunningDelayConsumer(topic, subName string) bool {
	_, ok := s.getDelayConsumer(topic, subName)
	return ok
}

// stopDelayConsumer returns an error representing a stop error.  If is OK, returns nil.
func (s *Nsq) stopDelayConsumer(topic string, subName string) {
	c, _ := s.getDelayConsumer(topic, subName)
	if c != nil {
		c.Consumer.Stop()
		log.Infof("close consumer success")
	}
	s.deleteDelayConsumer(topic, subName)
}

// ConsumerProcessorTopic is the topic of the subscription processor
func (s *Nsq) ConsumerProcessorTopic(confInfoList []*cc.ProcessorConfInfo) {
	defer routine.Recover()
	for i := range confInfoList {
		e, err := event.New(s.Opts, s.Handler, s.RedisPool, confInfoList[i])
		if err != nil {
			log.Errorf("make event err:%v", err)
			return
		}
		s.subscribeProcessor(e, confInfoList[i])
	}
}

func (s *Nsq) subscribeProcessor(e *event.Event, confInfo *cc.ProcessorConfInfo) {
	h := &Handler{}
	topic := topicutil.DelayTopic(confInfo.AppID, confInfo.TopicID)
	subName := topicutil.DelaySubName(confInfo.AppID, confInfo.TopicID)
	s.storeProcessorConsumeMap(confInfo.DispatcherID, h)
	go s.subscribe(topic, subName, e, h)
	log.Infof("subscribe confInfo:%+v", confInfo)
}

// isRunningProcessorConsumer returns true if the processor is running
func (s *Nsq) isRunningProcessorConsumer(dispatcherID string) bool {
	_, ok := s.getProcessorConsumeMap(dispatcherID)
	return ok
}

// stopRunningProcessorConsumer returns an error representing a stop error.  If is OK, returns nil.
func (s *Nsq) stopRunningProcessorConsumer(dispatcherID string) {
	c, _ := s.getProcessorConsumeMap(dispatcherID)
	if c != nil {
		c.Consumer.Stop()
		log.Infof("close consumer success")
		return
	}
}

// WatchProcessorTopic watch the topic of the processor
func (s *Nsq) WatchProcessorTopic(c *cc.ProcessorConfInfo, eventType int32) error {
	if eventType != cc.EventTypeDelete {
		if s.isRunningProcessorConsumer(c.DispatcherID) {
			s.stopRunningProcessorConsumer(c.DispatcherID)
		}
		e, err := event.New(s.Opts, s.Handler, s.RedisPool, c)
		if err != nil {
			log.Errorf("make event err:%v", err)
			return err
		}
		s.subscribeProcessor(e, c)
		return nil
	}
	if !s.isRunningProcessorConsumer(c.DispatcherID) {
		log.Infof("not running consumer conf:%+v", c)
		return nil
	}
	s.stopRunningProcessorConsumer(c.DispatcherID)
	s.deleteProcessorConsumeMap(c.DispatcherID)
	log.Infof("stop running consumer success info:%+v", c)
	return nil
}

// Stop stops the Nsq connector gracefully.
func (s *Nsq) Stop() error {
	s.delayConsumerMap.Range(func(key, value interface{}) bool {
		vv, ok := value.(*Handler)
		if !ok {
			log.Errorf("convert err:%v", value)
			return true
		}
		vv.Consumer.Stop()
		log.Infof("stop consumer:%+v success ", vv)
		return true
	})
	log.Infof("stop nsq consumer success")
	return nil
}
