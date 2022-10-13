package kafka

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/tencentmusic/evhub/internal/processor/define"

	"github.com/Shopify/sarama"
	"github.com/gomodule/redigo/redis"
	"github.com/tencentmusic/evhub/internal/processor/event"
	cc "github.com/tencentmusic/evhub/pkg/conf_connector"
	"github.com/tencentmusic/evhub/pkg/log"
	"github.com/tencentmusic/evhub/pkg/util/routine"
	"go.uber.org/ratelimit"
)

// ConsumerGroup is a consumer of Kafka
type ConsumerGroup struct {
	// a consumer group for Kafka topic
	GroupID string
	//  If is OK, representing auto commit.
	OffsetsAutoCommitEnable bool
	// flow control batch number for Kafka consumer group
	FlowControlBatchNum int
	// initial offset timeout for Kafka consumer group
	InitOffsetGetOffsetTimeout int64
	// initial offset period for Kafka consumer group
	InitOffsetGetOffsetPeroid int64
	// a processor of configuration
	confInfo *cc.ProcessorConfInfo
	// event handler
	Event *event.Event
	// redis pool
	RedisPool *redis.Pool
	//  a consumer group for Kafka topic
	ConsumerGroup sarama.ConsumerGroup
	// context cancel function for consumer group
	cancelFunc context.CancelFunc
	// the stop controller of flow control
	StopWg sync.WaitGroup
}

const (
	// KeyTopicGroupOffset is the key of the consumer group
	KeyTopicGroupOffset = "evhub:processor:%v:%v:%v"
	// CommitTimeout commit timeout
	CommitTimeout = 2
)

// TopicGroupOffsetKey returns the key of consumer group
func TopicGroupOffsetKey(topic string, group string, partition int32) string {
	return fmt.Sprintf(KeyTopicGroupOffset, topic, group, partition)
}

// Setup is the beginning of the processor
func (s *ConsumerGroup) Setup(session sarama.ConsumerGroupSession) error {
	log.Infof("session info:%v", session.Claims())
	return nil
}

// Cleanup is the end of the processor
func (s *ConsumerGroup) Cleanup(sarama.ConsumerGroupSession) error {
	log.Infof("session cleanup")
	return nil
}

// Msg is the message of the consumer group
type Msg struct {
	Message *sarama.ConsumerMessage
}

// ConsumeClaim is processing the message
func (s *ConsumerGroup) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	defer routine.Recover()
	msg := &Msg{}
	s.StopWg.Add(1)
	// flow control
	if s.confInfo.IsFlowControl {
		go s.flowControlConsumer(session, claim, msg)
	} else {
		go s.partitionConsumer(session, claim, msg)
	}
	log.Infof("message:%v", msg.Message)
	// listening rebalance
	<-session.Context().Done()
	log.Infof("consume claim success handle "+
		"rabalance topic:%v group id:%v partition:%v msg:%v message:%v",
		claim.Topic(), s.GroupID, claim.Partition(), msg, msg.Message)
	return nil
}

// flowControlConsumer is a flow control consumer
func (s *ConsumerGroup) flowControlConsumer(session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim, msg *Msg) {
	defer func() {
		routine.Recover()
		s.StopWg.Done()
	}()
	// rate limit consumer group
	rl := ratelimit.New(s.confInfo.FlowControlStrategy.QPS)
	wg := sync.WaitGroup{}
	// flow control batch number
	batchIndex := 0
	t := time.NewTimer(time.Second * time.Duration(CommitTimeout))
	var markMessageOffset int64
	var ok bool
	var message *sarama.ConsumerMessage
	for {
		select {
		case <-t.C:
			wg.Wait()
			if msg.Message != nil && msg.Message.Offset > markMessageOffset {
				markMessageOffset = msg.Message.Offset
				session.MarkMessage(msg.Message, "")
				if !s.OffsetsAutoCommitEnable {
					session.Commit()
				}
				log.Debugf("mark message partition:%v "+
					"offset:%v topic:%v", msg.Message.Partition,
					msg.Message.Offset, msg.Message.Topic)
			}
			batchIndex = 0
		case message, ok = <-claim.Messages():
			if !ok {
				log.Infof("handle report data chan success")
				goto Loop
			}
			msg.Message = message
			// flow control
			rl.Take()
			batchIndex++
			wg.Add(1)
			// handle message
			go s.handleConsumeMessage(*msg.Message, &wg)
			if batchIndex == s.FlowControlBatchNum {
				wg.Wait()
				markMessageOffset = msg.Message.Offset
				session.MarkMessage(msg.Message, "")
				if !s.OffsetsAutoCommitEnable {
					session.Commit()
				}
				log.Debugf("mark message partition:%v offset:%v"+
					" topic:%v", msg.Message.Partition, msg.Message.Offset,
					msg.Message.Topic)
				batchIndex = 0
			}
		}
		t.Stop()
		t = time.NewTimer(time.Second * time.Duration(CommitTimeout))
	}
Loop:
	// stop consumer
	log.Infof("consumer stop success")
}

// partitionConsumer is a partition consumer
// nolint
func (s *ConsumerGroup) partitionConsumer(session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim, msg *Msg) {
	defer routine.Recover()
	defer s.StopWg.Done()
	var ok bool
	var message *sarama.ConsumerMessage
	for {
		select {
		case message, ok = <-claim.Messages():
			if !ok {
				log.Infof("handle report data chan success")
				goto Loop
			}
			msg.Message = message
			log.Debugf("message claimed: key = %s, topic = %s partition = %v, "+
				"offset = %v value = %s, timestamp = %v",
				string(msg.Message.Key), msg.Message.Topic, msg.Message.Partition,
				msg.Message.Offset, string(msg.Message.Value), msg.Message.Timestamp)
			var buff bytes.Buffer
			buff.Write(msg.Message.Value)
			s.ConsumeMessage(buff.Bytes())
			log.Debugf("handle success %v %v %v", msg.Message.Topic,
				msg.Message.Partition, msg.Message.Offset)
			session.MarkMessage(msg.Message, "")
			if !s.OffsetsAutoCommitEnable {
				// manual commit
				session.Commit()
			}
			log.Debugf("handle success %v %v %v", msg.Message.Topic,
				msg.Message.Partition, msg.Message.Offset)
		}
	}
Loop:
	log.Infof("consumer stop success")
}

// handleConsumeMessage is processing the message
// nolint
func (s *ConsumerGroup) handleConsumeMessage(message sarama.ConsumerMessage, wg *sync.WaitGroup) {
	defer func() {
		routine.Recover()
		wg.Done()
	}()
	log.Debugf("message claimed: key = %s, "+
		"topic = %s partition = %v, offset = %v value = %s, timestamp = %v",
		string(message.Key), message.Topic, message.Partition,
		message.Offset, string(message.Value), message.Timestamp)
	s.ConsumeMessage(message.Value)
}

// ConsumeMessage consumer messages
func (s *ConsumerGroup) ConsumeMessage(msg []byte) {
	if err := s.Event.HandleEvent(msg, define.EventTypeExternal); err != nil {
		log.Infof("handle event fn err: %v", err)
	}
}

// Stop stops the consumer group gracefully.
func (s *ConsumerGroup) Stop() error {
	s.cancelFunc()
	s.StopWg.Wait()
	s.ConsumerGroup.Close()
	log.Infof("stop consumer group success kafka consumer:%v", s)
	return nil
}
