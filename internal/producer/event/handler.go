package event

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/tencentmusic/evhub/internal/producer/define"

	"github.com/tencentmusic/evhub/internal/producer/connector"
	"github.com/tencentmusic/evhub/pkg/bizno"
	"github.com/tencentmusic/evhub/pkg/gen/proto/comm"
	"github.com/tencentmusic/evhub/pkg/log"
	"github.com/tencentmusic/evhub/pkg/util/topicutil"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Handler is the handler for events
type Handler struct {
	// DelayConnector is the name of the delay connector
	DelayConnector connector.Connector
	// RealTimeConnector is the name of the real time connector
	RealTimeConnector connector.Connector
	// BizNO is the name of the biz number
	BizNO bizno.Pool
}

var cronParser = cron.NewParser(cron.Second |
	cron.Minute | cron.Hour | cron.Dom | cron.Month |
	cron.DowOptional | cron.Descriptor)

// SendMsg is the stores that send  events
func (s *Handler) SendMsg(ctx context.Context, brokerTopic string, delayTime time.Duration, orderingKey string, msg []byte) error {
	if delayTime.Milliseconds() == 0 {
		return s.RealTimeConnector.SendMsg(ctx, brokerTopic, orderingKey, msg)
	}
	return s.DelayConnector.SendDelayMsg(ctx, brokerTopic, orderingKey, msg, delayTime)
}

// handleMsg is used to handle a message
func (s *Handler) handleMsg(ctx context.Context, in *define.ReportReq, eventID string) error {
	b, err := s.makeEvhubMsg(in, eventID)
	if err != nil {
		return nil
	}
	// creates topic prefix for the event
	topicPrefix, delayTime := s.makeTopicPrefix(in)
	log.With(ctx).Debugf("topic prefix:%v delay time:%v", topicPrefix, delayTime)
	if err = s.SendMsg(ctx, BrokerTopicKey(topicPrefix, in.Event.AppId, in.Event.TopicId),
		delayTime, s.makePartitionKey(in), b); err != nil {
		return err
	}
	return nil
}

// makeEvhubMsg creates a new EvhubMsg
func (s *Handler) makeEvhubMsg(in *define.ReportReq, eventID string) ([]byte, error) {
	msg := &comm.EvhubMsg{
		EventId:     eventID,
		Event:       in.Event,
		TriggerInfo: &comm.EventTriggerInfo{},
	}
	msg.Status = comm.EventStatus_EVENT_STATUS_INIT
	llEventCreateTime := time.Now()
	if msg.Event.EventComm != nil && msg.Event.EventComm.EventSource != nil {
		llEventCreateTime = msg.Event.EventComm.EventSource.CreateTime.AsTime()
	}

	if llEventCreateTime.UnixNano() == 0 {
		msg.Event.EventComm.EventSource.CreateTime = timestamppb.New(time.Now())
	}

	msg.TriggerInfo.Trigger = in.Trigger
	// marshal message
	msgByte, err := proto.Marshal(msg)
	if err != nil {
		return nil, err
	}
	return msgByte, nil
}

// makeTopicPrefix returns the topic prefix
func (s *Handler) makeTopicPrefix(req *define.ReportReq) (string, time.Duration) {
	topicPrefix := topicutil.TopicPrefix
	delayTime := durationpb.New(time.Duration(0))
	triggerType := req.Trigger.TriggerType
	// initial repeat beginning time
	repeatBeginTime := time.Now()
	if req.Trigger != nil && req.Trigger.RepeatTrigger != nil {
		repeatBeginTime = req.Trigger.RepeatTrigger.BeginTime.AsTime()
	}

	switch triggerType {
	// delay event
	case comm.EventTriggerType_EVENT_TRIGGER_TYPE_DELAY:
		delayTime = req.Trigger.DelayTrigger.DelayTime
		if delayTime.AsDuration().Milliseconds() > 0 {
			topicPrefix = topicutil.DelayTopicPrefix
		}
	// repeat event
	case comm.EventTriggerType_EVENT_TRIGGER_TYPE_REPEAT:
		if repeatBeginTime.UnixNano() > (time.Now().UnixNano()) {
			topicPrefix = topicutil.DelayTopicPrefix
			delayTime = durationpb.New(time.Duration(repeatBeginTime.UnixNano() - (time.Now().UnixNano())))
		}
	// delay event
	case comm.EventTriggerType_EVENT_TRIGGER_TYPE_DELAY_PROCESS, comm.EventTriggerType_EVENT_TRIGGER_TYPE_REAL_TIME:
		return topicPrefix, delayTime.AsDuration()
	default:
	}
	return topicPrefix, delayTime.AsDuration()
}

func (s *Handler) makePartitionKey(in *define.ReportReq) string {
	if in.Option == nil {
		return ""
	}
	return in.Option.OrderingKey
}

// handleForkEvhubMsg handles evhub message
func (s *Handler) handleForkEvhubMsg(ctx context.Context,
	req *define.InternalHandlerReq) (string, time.Duration, error) {
	topicPrefix := topicutil.TopicPrefix
	// initialize delay time
	delayTime := time.Duration(0)
	if req.EvhubMsg.DispatcherInfo != nil && req.EvhubMsg.DispatcherInfo.RetryTimes > 0 {
		delayTime = req.EvhubMsg.TriggerInfo.Trigger.DelayTrigger.DelayTime.AsDuration()
		if delayTime.Milliseconds() > 0 {
			topicPrefix = topicutil.DelayTopicPrefix
		}
		return topicPrefix, delayTime, nil
	}

	switch {
	// repeat events
	case req.EvhubMsg.TriggerInfo.
		Trigger.TriggerType == comm.EventTriggerType_EVENT_TRIGGER_TYPE_REPEAT &&
		req.EvhubMsg.Status == comm.EventStatus_EVENT_STATUS_PROCESS_SUCC:
		return s.repeatEventType(req)
	// delay events
	case req.EvhubMsg.TriggerInfo.
		Trigger.TriggerType == comm.EventTriggerType_EVENT_TRIGGER_TYPE_DELAY:
		delayTime = req.EvhubMsg.TriggerInfo.Trigger.DelayTrigger.DelayTime.AsDuration()
		if delayTime.Milliseconds() > 0 {
			topicPrefix = topicutil.DelayTopicPrefix
		}

	default:
		log.With(ctx).Debugf("delay event to real time event")
	}
	return topicPrefix, delayTime, nil
}

// repeatEventType returns topic prefix and delay time for repeat event
func (s *Handler) repeatEventType(req *define.InternalHandlerReq) (string, time.Duration, error) {
	delayTime := time.Duration(0)
	topicPrefix := topicutil.TopicPrefix
	switch req.EvhubMsg.TriggerInfo.Trigger.RepeatTrigger.PulserType {
	case comm.EventPulserType_EVENT_PULSER_TYPE_BY_INTERVAL:
		delayTime = req.EvhubMsg.TriggerInfo.Trigger.RepeatTrigger.RegularPulser.Interval.AsDuration()
	case comm.EventPulserType_EVENT_PULSER_TYPE_BY_CRONTAB:
		sc, err := cronParser.Parse(req.EvhubMsg.
			TriggerInfo.Trigger.RepeatTrigger.
			CrontabPulser.Crontab)
		if err != nil {
			log.Errorf("crontab parse err: %v req:%v", err, req)
			return "", 0, nil
		}
		nextTime := sc.Next(time.Now())
		delayTime = time.Until(nextTime)
	default:
		return topicPrefix, delayTime, fmt.Errorf("eventPulserType Err")
	}
	if delayTime.Milliseconds() > 0 {
		topicPrefix = topicutil.DelayTopicPrefix
	}
	return topicPrefix, delayTime, nil
}

// verifyFork is used to verify the request
func (s *Handler) verifyFork(req *define.InternalHandlerReq) error {
	if req == nil || req.EvhubMsg.Event == nil {
		return errors.New("req is nil")
	}
	appID := req.EvhubMsg.Event.AppId
	topicID := req.EvhubMsg.Event.TopicId

	if appID == "" || topicID == "" {
		return errors.New("request param error app id or topic id is null")
	}
	return nil
}

// verifyReport is used to verify the request
func (s *Handler) verifyReport(req *define.ReportReq) error {
	if req == nil || req.Event == nil {
		return errors.New("req is nil")
	}
	appID := req.Event.AppId
	topicID := req.Event.TopicId

	if appID == "" || topicID == "" {
		return errors.New("request param error app id or topic id is null")
	}
	return nil
}
