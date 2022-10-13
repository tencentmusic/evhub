package event

import (
	"context"
	"fmt"
	"time"

	"github.com/tencentmusic/evhub/internal/processor/define"

	redigo "github.com/gomodule/redigo/redis"

	"github.com/tencentmusic/evhub/internal/processor/options"
	cc "github.com/tencentmusic/evhub/pkg/conf_connector"
	"github.com/tencentmusic/evhub/pkg/errcode"
	"github.com/tencentmusic/evhub/pkg/gen/proto/comm"
	"github.com/tencentmusic/evhub/pkg/handler"
	"github.com/tencentmusic/evhub/pkg/log"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
)

// Event include loops, delays, transactions, and real-time events
type Event struct {
	// opts is the name of the processor options
	opts *options.Options
	// handler is the name of the processor handler
	handler *Handler
	// redisPool is the name of the redis pool
	redisPool *redigo.Pool
	// confInfo is the name of the processor configuration
	confInfo *cc.ProcessorConfInfo
}

// New creates an Event
func New(opts *options.Options, h *handler.Handler,
	redisPool *redigo.Pool, confInfo *cc.ProcessorConfInfo) (*Event, error) {
	return &Event{
		opts: opts,
		handler: &Handler{
			handler:  h,
			opts:     opts,
			confInfo: confInfo,
		},
		confInfo:  confInfo,
		redisPool: redisPool,
	}, nil
}

// getDispatchEventID returns the key for the dispatcher event
func getDispatchEventID(msg *comm.EvhubMsg) string {
	return fmt.Sprintf("%v_%v_%v_%v", msg.DispatcherInfo.DispatcherId,
		msg.EventId, msg.TriggerInfo.RepeatTimes, msg.DispatcherInfo.RetryTimes)
}

// HandleEvent handler events for the dispatcher
func (s *Event) HandleEvent(msg []byte, eventType define.EventType) error {
	if eventType == define.EventTypeExternal {
		return s.HandleEventExternal(msg)
	}
	if eventType == define.EventTypeInternal {
		return s.HandleEventInternal(msg)
	}
	return nil
}

// HandleEventExternal handler external events for the dispatcher
func (s *Event) HandleEventExternal(data []byte) error {
	// verify
	msg, err := s.verify(data)
	if err != nil {
		return fmt.Errorf("verify err:%v data:%v", err, string(data))
	}
	// log dyeing
	ctx := log.SetDye(context.Background(), msg.Event.AppId, msg.Event.TopicId,
		msg.EventId, msg.DispatcherInfo.RetryTimes, msg.TriggerInfo.RepeatTimes)

	// repeat duplicate
	if err = s.repeatDuplicate(ctx, msg); err != nil {
		log.With(ctx).Infof("repeat duplicate evhub msg:%+v err:%v", msg, err)
		return nil
	}
	// filter tag
	if !s.filterTag(msg) {
		return nil
	}
	log.With(ctx).Debugf("evhub msg:%+v", msg)
	// delay processor
	if s.isDelayProcess(ctx, msg) {
		log.With(ctx).Debugf("evhub msg delay process evhub msg:%v delay process key:%v event dispatcher id:%v",
			msg, s.confInfo.DelayProcessKey, s.confInfo.DispatcherID)
		return nil
	}
	// send request
	passBack, err := s.handler.Send(ctx, msg)
	msg.DispatcherInfo.PassBack = nil
	if err != nil {
		log.With(ctx).Infof("processor event err:%v", err)
		if !s.isNeedRetryProcess(ctx, msg, passBack, errcode.ConvertRet(err).Code) {
			log.With(ctx).Infof("processor send err:%v evhub msg:%v s:%v", err, msg, s)
		} else {
			msg.DispatcherInfo.RetryTimes--
		}
	}
	// repeat processing
	s.repeatProcess(ctx, msg, errcode.ConvertRet(err).Code)
	s.repeatDuplicateCommit(ctx, msg)
	return nil
}

// HandleEventInternal handler internal events for the dispatcher
func (s *Event) HandleEventInternal(data []byte) error {
	var msg comm.EvhubMsg
	// unmarshal message
	if err := proto.Unmarshal(data, &msg); err != nil {
		return err
	}
	if msg.Status != comm.EventStatus_EVENT_STATUS_TX_CALLBACK_IN_DELAY_QUEUE {
		if msg.TriggerInfo != nil && msg.TriggerInfo.Trigger != nil && msg.TriggerInfo.Trigger.DelayTrigger != nil {
			msg.TriggerInfo.Trigger.DelayTrigger.DelayTime = durationpb.New(time.Duration(0))
		}
		msg.Status = comm.EventStatus_EVENT_STATUS_DELAY_RIGGER
	}
	s.handler.InternalHandlerSuccess(context.Background(), &msg)
	return nil
}

// verify event parameters
func (s *Event) verify(data []byte) (*comm.EvhubMsg, error) {
	var msg comm.EvhubMsg
	if err := proto.Unmarshal(data, &msg); err != nil {
		log.Errorf("unmarshal evhub msg:%+v err=%v", &msg, err)
		return nil, err
	}
	// not processing
	if msg.DispatcherInfo != nil && msg.DispatcherInfo.DispatcherId != "" &&
		msg.DispatcherInfo.DispatcherId != s.confInfo.DispatcherID {
		return &msg, fmt.Errorf("not need producer EvhubMsg eventDispatcherId:%v eventDispatcherID:%v",
			msg.DispatcherInfo.DispatcherId, s.confInfo.DispatcherID)
	}
	// initialize dispatcher id
	if msg.DispatcherInfo == nil || msg.DispatcherInfo.DispatcherId == "" {
		msg.DispatcherInfo = &comm.EventDispatcherInfo{
			DispatcherId: s.confInfo.DispatcherID,
		}
	}

	return &msg, nil
}

// isDelayProcess returns true if the event is delay processing
func (s *Event) isDelayProcess(ctx context.Context, msg *comm.EvhubMsg) bool {
	// not delay processing
	if msg.TriggerInfo.Trigger.TriggerType != comm.EventTriggerType_EVENT_TRIGGER_TYPE_DELAY_PROCESS {
		return false
	}
	if !(s.confInfo.DelayProcessKey != "" && len(msg.TriggerInfo.Trigger.DelayProcessTriggerList) > 0) {
		return false
	}
	// delay processing
	for _, eventDelayProcessTrigger := range msg.TriggerInfo.Trigger.DelayProcessTriggerList {
		if eventDelayProcessTrigger.Key != s.confInfo.DelayProcessKey {
			continue
		}
		msg.DispatcherInfo.DispatcherId = s.confInfo.DispatcherID
		msg.TriggerInfo.Trigger.TriggerType = comm.EventTriggerType_EVENT_TRIGGER_TYPE_DELAY
		msg.TriggerInfo.Trigger.DelayTrigger.DelayTime = eventDelayProcessTrigger.DelayTime
		s.handler.InternalHandlerSuccess(ctx, msg)
		return true
	}
	log.With(ctx).Errorf("evhub msg err evhub msg:%+v delay process key:%v event dispatcher id:%v",
		msg, s.confInfo.DelayProcessKey, s.confInfo.DispatcherID)
	return true
}

// repeatProcess loop event processing
func (s *Event) repeatProcess(ctx context.Context, msg *comm.EvhubMsg, ret int32) {
	// stop processing for this event
	if !(msg.TriggerInfo.Trigger.TriggerType ==
		comm.EventTriggerType_EVENT_TRIGGER_TYPE_REPEAT &&
		msg.DispatcherInfo.RetryTimes == 0 &&
		ret != errcode.EventCancel) {
		log.With(ctx).Debugf("no repeat process evhub msg:%+v", msg)
		return
	}
	needRepeat := false
	switch msg.TriggerInfo.Trigger.RepeatTrigger.EndType {
	// end timestamp
	case comm.EventRepeatEndType_EVENT_REPEAT_END_TYPE_END_BY_TIMESTAMP:
		if msg.TriggerInfo.Trigger.RepeatTrigger.RepeatEnd.EndTime.AsTime().UnixNano() == 0 ||
			msg.TriggerInfo.Trigger.RepeatTrigger.RepeatEnd.EndTime.AsTime().UnixNano()/1e6 >
				(time.Now().UnixNano()/1e6) {
			needRepeat = true
		}
	//	end count
	case comm.EventRepeatEndType_EVENT_REPEAT_END_TYPE_END_BY_COUNT:
		if msg.TriggerInfo.Trigger.RepeatTrigger.RepeatEnd.RepeatTimes == 0 ||
			msg.TriggerInfo.RepeatTimes < (msg.TriggerInfo.Trigger.RepeatTrigger.RepeatEnd.RepeatTimes-1) {
			needRepeat = true
		}
	default:
		log.With(ctx).Errorf("event repeat end type err!")
		return
	}
	if !needRepeat {
		log.With(ctx).Debugf("no repeat process evhub msg:%+v", msg)
		return
	}
	msg.TriggerInfo.RepeatTimes++
	msg.DispatcherInfo.RetryTimes = 0
	msg.Status = comm.EventStatus_EVENT_STATUS_PROCESS_SUCC
	// internal handling for events
	s.handler.InternalHandlerSuccess(ctx, msg)
	log.With(ctx).Debugf("need repeat repeat times:%v evhub msg:%+v", msg.TriggerInfo.RepeatTimes, msg)
	msg.TriggerInfo.RepeatTimes--
}

// isDelayProcess returns true if the event need to be processed
func (s *Event) isNeedRetryProcess(ctx context.Context, msg *comm.EvhubMsg, passBack []byte, ret int32) bool {
	// not need to retry
	if s.confInfo.RetryStrategy.RetryCount <= 0 || ret == errcode.EventCancel {
		return false
	}
	retryTime := int(msg.DispatcherInfo.RetryTimes)
	log.With(ctx).Debugf("need retry retrytimes:%v "+
		"retrymaxTimes:%v evhub msg:%v", retryTime, s.confInfo.RetryStrategy.RetryCount, msg)
	if s.confInfo.RetryStrategy.RetryCount <= retryTime+1 {
		return false
	}
	msg.DispatcherInfo.RetryTimes++
	msg.DispatcherInfo.DispatcherId = s.confInfo.DispatcherID
	msg.DispatcherInfo.PassBack = passBack
	if msg.TriggerInfo == nil {
		msg.TriggerInfo = &comm.EventTriggerInfo{}
	}
	if msg.TriggerInfo.Trigger == nil {
		msg.TriggerInfo.Trigger = &comm.EventTrigger{}
	}
	if msg.TriggerInfo.Trigger.DelayTrigger == nil {
		msg.TriggerInfo.Trigger.DelayTrigger = &comm.EventDelayTrigger{}
	}
	msg.TriggerInfo.Trigger.DelayTrigger.DelayTime = durationpb.New(time.Millisecond *
		time.Duration(s.confInfo.RetryStrategy.RetryInterval))
	s.handler.InternalHandlerSuccess(ctx, msg)
	return true
}

// repeatDuplicateCommit loops event is duplicate commit
func (s *Event) repeatDuplicateCommit(ctx context.Context, msg *comm.EvhubMsg) {
	if msg.TriggerInfo.Trigger.TriggerType !=
		comm.EventTriggerType_EVENT_TRIGGER_TYPE_REPEAT {
		return
	}
	conn := s.redisPool.Get()
	defer conn.Close()
	reply, err := conn.Do("SET", define.TopicRepeatDuplicateKey(getDispatchEventID(msg)),
		define.RepeatDuplicateCommitted, "EX", define.RepeatDuplicateTimeOut)
	if err != nil {
		log.With(ctx).Errorf("redis err:%v", err)
		return
	}
	if reply != "OK" {
		log.Errorf("repeat duplicate EvhubMsg:%+v", msg)
		return
	}
}

// filterTag filter tag for event
func (s *Event) filterTag(msg *comm.EvhubMsg) bool {
	log.Debugf("msg:%+v confInfo:%+v", msg, s.confInfo)
	if s.confInfo.FilterTag == nil {
		return true
	}
	log.Debugf("msg: %+v", msg)
	if !(msg.Event != nil && msg.Event.EventComm != nil && msg.Event.EventComm.EventTag != nil) {
		return false
	}
	for _, tag := range msg.Event.EventComm.EventTag.TagList {
		if _, ok := s.confInfo.FilterTag[tag]; ok {
			return true
		}
	}
	return false
}

// repeatDuplicate
func (s *Event) repeatDuplicate(ctx context.Context, msg *comm.EvhubMsg) error {
	if msg.TriggerInfo.Trigger.TriggerType !=
		comm.EventTriggerType_EVENT_TRIGGER_TYPE_REPEAT {
		return nil
	}
	conn := s.redisPool.Get()
	defer conn.Close()
	script := redigo.NewScript(1, define.RepeatDuplicateScript)
	evals, err := script.Do(conn, define.TopicRepeatDuplicateKey(getDispatchEventID(msg)), define.RepeatDuplicateTimeOut)
	if err != nil {
		log.With(ctx).Errorf("redis err:%v", err)
		return nil
	}
	log.With(ctx).Debugf("repeat duplicate r:%v", evals)
	vals, ok := evals.([]interface{})
	if !ok {
		log.With(ctx).Errorf("convert evals err")
		return nil
	}
	log.With(ctx).Infof("store offset topic repeat duplicate key:%v repeat duplicate timeOut:%v",
		define.TopicRepeatDuplicateKey(getDispatchEventID(msg)), define.RepeatDuplicateTimeOut)
	if len(vals) == 0 {
		log.With(ctx).Errorf("redis err:%v", err)
		return nil
	}
	r := string(vals[0].([]byte))
	log.With(ctx).Debugf("repeat duplicate script r:%v", r)
	switch r {
	case define.RepeatDuplicatePre:
		return nil
	case define.RepeatDuplicateCommitted:
		return fmt.Errorf("repeat duplicate")
	case define.RepeatDuplicatePreFirst:
		return nil
	case define.RepeatDuplicateGetErr:
		log.Errorf("repeat duplicate err:%v", r)
		return nil
	case define.RepeatDuplicateSetErr:
		log.Errorf("repeat duplicate err:%v", r)
		return nil
	default:
		log.Errorf("repeat duplicate err:%v")
		return nil
	}
}

// Stop stops the Kafka connector gracefully.
func (s *Event) Stop() error {
	return nil
}
