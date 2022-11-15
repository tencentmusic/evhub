package event

import (
	"context"
	"fmt"

	"github.com/tencentmusic/evhub/internal/producer/connector"
	"github.com/tencentmusic/evhub/internal/producer/define"
	"github.com/tencentmusic/evhub/internal/producer/options"
	"github.com/tencentmusic/evhub/pkg/bizno"
	cc "github.com/tencentmusic/evhub/pkg/conf_connector"
	"github.com/tencentmusic/evhub/pkg/errcode"
	"github.com/tencentmusic/evhub/pkg/gen/proto/comm"
	"github.com/tencentmusic/evhub/pkg/log"
	"google.golang.org/protobuf/proto"
)

// Event include loops, delays, transactions, and real-time events
type Event struct {
	// opts is the name of the producer configuration
	opts *options.Options
	// transaction is name of the transaction
	transaction *Transaction
	// handler is the name of the handler
	handler *Handler
	// producerConfConnector is the name of the producer configuration
	producerConfConnector cc.ProducerConfConnector
}

const (
	KeyBrokerTopic = "%v%v_%v"
)

// BrokerTopicKey returns the broker topic
func BrokerTopicKey(topicPrefix string, appID string, topicID string) string {
	return fmt.Sprintf(KeyBrokerTopic, topicPrefix, appID, topicID)
}

// New creates an Event
func New(opts *options.Options) (*Event, error) {
	// initialize biz number generator
	bizNO, err := bizno.NewPool(&opts.BizNO)
	if err != nil {
		return nil, err
	}
	// initialize real time connector
	realTimeConnector, err := connector.New(&opts.RealTimeConnectorConfig)
	if err != nil {
		return nil, err
	}
	// initialize delay connector
	delayConnector, err := connector.New(&opts.DelayConnectorConfig)
	if err != nil {
		return nil, err
	}
	handler := &Handler{
		DelayConnector:    delayConnector,
		RealTimeConnector: realTimeConnector,
		BizNO:             bizNO,
	}
	// initialize producer conf connector
	producerConfConnector, err := cc.NewProducerConfConnector(opts.ConfConnectorConfig)
	if err != nil {
		return nil, err
	}
	confInfoList, err := producerConfConnector.GetAllConfInfo()
	if err != nil {
		return nil, err
	}
	tx, err := NewTransaction(&opts.TxConfig, handler, confInfoList)
	if err != nil {
		return nil, err
	}
	if err = tx.Run(); err != nil {
		return nil, err
	}
	if err = producerConfConnector.SetWatch(tx.WatchProducerConf); err != nil {
		return nil, err
	}
	e := &Event{
		opts:                  opts,
		handler:               handler,
		transaction:           tx,
		producerConfConnector: producerConfConnector,
	}

	return e, nil
}

// Report is used to report events to the server
func (s *Event) Report(ctx context.Context, req interface{}) (interface{}, error) {
	in, ok := req.(*define.ReportReq)
	if !ok {
		return nil, fmt.Errorf("convert request error: %v", in)
	}
	eventID, err := s.handler.BizNO.NextID()
	if err != nil {
		log.Errorf("generator id err:%v", err)
		return &define.ReportRsp{
			Ret: &comm.Result{
				Code: errcode.SystemErr,
				Msg:  "system error",
			},
		}, nil
	}
	r := &define.ReportRsp{
		EventID: eventID,
	}
	// verify
	if err = s.handler.verifyReport(in); err != nil {
		log.Infof("verify err:%v", err)
		r.Ret = &comm.Result{
			Code: errcode.CommParamInvalid,
			Msg:  err.Error(),
		}
		return r, nil
	}
	appID := in.Event.AppId
	topicID := in.Event.TopicId
	var retryTimes, repeatTimes uint32
	ctx = log.SetDye(ctx, appID, topicID, eventID, retryTimes, repeatTimes)

	if err = s.handler.handleMsg(ctx, in, eventID); err != nil {
		log.With(ctx).Infof("handle msg err err:%v", err)
		r.Ret = &comm.Result{
			Code: errcode.ProduceFail,
			Msg:  err.Error(),
		}
		return r, nil
	}
	return r, nil
}

// Prepare is used to prepare a transaction message to the server
func (s *Event) Prepare(ctx context.Context, req interface{}) (interface{}, error) {
	in, ok := req.(*define.PrepareReq)
	if !ok {
		return nil, fmt.Errorf("")
	}
	eventID, err := s.transaction.Prepare(ctx, in)
	if err != nil {
		return &define.PrepareRsp{
			Ret: errcode.ConvertRet(err),
			Tx: &comm.Tx{
				EventId: eventID,
			},
		}, nil
	}
	return &define.PrepareRsp{
		Tx: &comm.Tx{
			EventId: eventID,
			Status:  comm.TxStatus_TX_STATUS_PREPARE,
		},
	}, nil
}

// Commit is used to commit a transaction message to the server
func (s *Event) Commit(ctx context.Context, req interface{}) (interface{}, error) {
	in, ok := req.(*define.CommitReq)
	if !ok {
		return nil, fmt.Errorf("")
	}
	tx, err := s.transaction.Commit(ctx, in)
	if err != nil {
		return &define.CommitRsp{
			Ret: errcode.ConvertRet(err),
			Tx:  tx,
		}, nil
	}
	return &define.CommitRsp{
		Tx: tx,
	}, nil
}

// Rollback is used to roll back a transaction message to the server
func (s *Event) Rollback(ctx context.Context, req interface{}) (interface{}, error) {
	in, ok := req.(*define.RollbackReq)
	if !ok {
		return nil, fmt.Errorf("")
	}
	tx, err := s.transaction.Rollback(ctx, in)
	if err != nil {
		return &define.RollbackRsp{
			Ret: errcode.ConvertRet(err),
			Tx:  tx,
		}, nil
	}
	return &define.RollbackRsp{
		Tx: tx,
	}, nil
}

// InternalHandler is used to handle internal messages
func (s *Event) InternalHandler(ctx context.Context, req interface{}) (interface{}, error) {
	in, ok := req.(*define.InternalHandlerReq)
	if !ok {
		return nil, fmt.Errorf("convert err")
	}
	r := &define.InternalHandlerRsp{}
	appID := in.EvhubMsg.Event.AppId
	topicID := in.EvhubMsg.Event.TopicId
	eventID := in.EvhubMsg.EventId
	var retryTimes uint32
	if in.EvhubMsg.DispatcherInfo != nil {
		retryTimes = in.EvhubMsg.DispatcherInfo.RetryTimes
	}
	var repeatTimes uint32
	if in.EvhubMsg.TriggerInfo != nil {
		repeatTimes = in.EvhubMsg.TriggerInfo.RepeatTimes
	}
	ctx = log.SetDye(ctx, appID, topicID, eventID, retryTimes, repeatTimes)
	if in.EvhubMsg.Status == comm.EventStatus_EVENT_STATUS_TX_CALLBACK_IN_DELAY_QUEUE {
		log.With(ctx).Debugf("tx callback msg%+v", req)
		if err := s.transaction.CallBackInternal(ctx, &define.TxEventMsg{
			AppID:   appID,
			TopicID: topicID,
			EventID: eventID,
		}); err != nil {
			r.Ret = errcode.ConvertRet(err)
			return r, nil
		}
		return r, nil
	}

	topicPrefix, delayTime, err := s.handler.handleForkEvhubMsg(ctx, in)
	if err != nil {
		log.With(ctx).Errorf("makeForkEvhubMsg err:%v req:%v", err, req)
		return nil, err
	}

	if topicPrefix == "" {
		return &define.InternalHandlerRsp{}, nil
	}

	msgByte, err := proto.Marshal(in.EvhubMsg)
	if err != nil {
		return nil, err
	}
	if err = s.handler.SendMsg(ctx,
		BrokerTopicKey(topicPrefix, in.EvhubMsg.Event.AppId,
			in.EvhubMsg.Event.TopicId), delayTime, "", msgByte); err != nil {
		log.With(ctx).Errorf("senMsg err err:%v req:%v", err, req)
		return nil, err
	}
	return &define.InternalHandlerRsp{}, nil
}

// Stop stops the event gracefully.
func (s *Event) Stop() error {
	if err := s.transaction.Stop(); err != nil {
		return err
	}
	return nil
}
