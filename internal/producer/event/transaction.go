package event

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/panjf2000/ants/v2"

	"github.com/tencentmusic/evhub/internal/producer/define"
	"github.com/tencentmusic/evhub/pkg/util/addrutil"

	"github.com/tencentmusic/evhub/internal/producer/options"

	"github.com/tencentmusic/evhub/internal/producer/event/model"
	"github.com/tencentmusic/evhub/pkg/gen/proto/comm"
	eh_tx "github.com/tencentmusic/evhub/pkg/gen/proto/transaction"
	"google.golang.org/protobuf/types/known/timestamppb"

	cc "github.com/tencentmusic/evhub/pkg/conf_connector"
	"github.com/tencentmusic/evhub/pkg/grpc/codes"
	"github.com/tencentmusic/evhub/pkg/util/topicutil"

	"google.golang.org/protobuf/proto"

	"gorm.io/gorm"

	"github.com/tencentmusic/evhub/pkg/errcode"
	"github.com/tencentmusic/evhub/pkg/grpc"
	"github.com/tencentmusic/evhub/pkg/log"
	"github.com/tencentmusic/evhub/pkg/routine_pool"
	"github.com/tencentmusic/evhub/pkg/util/routine"
)

type CbCfType int

const (
	// TxStatusRunning TxStatus
	TxStatusRunning comm.TxStatus = 0
	// TxStatusPrepare
	TxStatusPrepare comm.TxStatus = 1
	// TxStatusCommit
	TxStatusCommit comm.TxStatus = 2
	// TxStatusRollback
	TxStatusRollback comm.TxStatus = 3
	// TxStatusSuccess
	TxStatusSuccess comm.TxStatus = 4

	// RetryCount is the number of times
	RetryCount = 20

	// AsyncReportEvent
	AsyncReportEvent = 0
	// AsyncCallBack
	AsyncCallBack = 1

	// CbHeartbeatTimeOut
	CbHeartbeatTimeOut = 60
	// CfHeartbeatTimeOut
	CfHeartbeatTimeOut = 60

	// CbExpire
	CbExpire = -60 * 5

	// TimeStrFormat
	TimeStrFormat = "2006-01-02 15:04:05"

	// PeriodLimit
	PeriodLimit = 10

	// CbCfTypeConfigParam
	CbCfTypeConfigParam CbCfType = 1
	// CbCfTypeConfigPlatform
	CbCfTypeConfigPlatform CbCfType = 2

	// CbCfTypeConfig
	DefaultCbTimeout = 5000
	// DefaultCbInterval
	DefaultCbInterval = 1000 * 60 * 2
)

// Transaction is a way of handling an event
type Transaction struct {
	// db is the database engine
	db *DB
	// readDB is the read database engine
	readDB *DB
	// handler is the name of the event handler
	handler *Handler
	// routinePool is the name of the routine pool
	routinePool *routine_pool.Pool
	// cbHeartbeat is the name of callback heartbeat
	cbHeartbeat *time.Ticker
	// cfHeartbeat is the name of commit failed heartbeat
	cfHeartbeat *time.Ticker
	// producerConfMap is the name of the producer configuration
	producerConfMap sync.Map
}

// NewTransaction creates a new transaction
func NewTransaction(txConfig *options.TransactionConfig,
	handler *Handler, confInfoList []*cc.ProducerConfInfo) (*Transaction, error) {
	// initialize read database
	readDB, err := NewDB(txConfig.ReadDBConfig)
	if err != nil {
		return nil, err
	}
	// initialize write database
	writeDB, err := NewDB(txConfig.WriteDBConfig)
	if err != nil {
		return nil, err
	}
	tx := &Transaction{
		db:      writeDB,
		readDB:  readDB,
		handler: handler,
	}
	// initialize routine pool
	tx.routinePool, err = routine_pool.New(txConfig.RoutinePoolSize, tx.AsyncReportEvent,
		ants.WithNonblocking(txConfig.Nonblocking))
	if err != nil {
		return nil, err
	}
	tx.SetProducerConfList(confInfoList)
	return tx, err
}

// KeyProducerConf returns the key of producer group
func KeyProducerConf(appID, topicID string) string {
	return fmt.Sprintf("%v_%v", appID, topicID)
}

// setProducerConf stores the producer configuration
func (s *Transaction) setProducerConf(appID, topicID string, conf *cc.ProducerConfInfo) {
	s.producerConfMap.Store(KeyProducerConf(appID, topicID), conf)
}

// deleteProducerConf delete the producer configuration
func (s *Transaction) deleteProducerConf(appID, topicID string) {
	s.producerConfMap.Delete(KeyProducerConf(appID, topicID))
}

// getProducerConf get the producer configuration
func (s *Transaction) getProducerConf(appID, topicID string) (*cc.ProducerConfInfo, bool) {
	v, ok := s.producerConfMap.Load(KeyProducerConf(appID, topicID))
	if !ok {
		return nil, ok
	}
	vv, ok := v.(*cc.ProducerConfInfo)
	if !ok {
		return nil, ok
	}
	return vv, true
}

// Run is used to run the transaction
func (s *Transaction) Run() error {
	routine.GoWithRecovery(s.periodCallBack)
	routine.GoWithRecovery(s.periodCommitFail)
	return nil
}

// SetProducerConfList is used to set the list of producers
func (s *Transaction) SetProducerConfList(confInfoList []*cc.ProducerConfInfo) {
	for _, confInfo := range confInfoList {
		s.setProducerConf(confInfo.AppID, confInfo.TopicID, confInfo)
	}
}

// WatchProducerConf is used to watch producer configuration
func (s *Transaction) WatchProducerConf(c *cc.ProducerConfInfo, eventType int32) error {
	if eventType == cc.EventTypeDelete {
		s.deleteProducerConf(c.AppID, c.TopicID)
		return nil
	}
	s.setProducerConf(c.AppID, c.TopicID, c)
	return nil
}

// makeCallbackReq creates a callback request for the given tx event
func (s *Transaction) makeCallbackReq(ctx context.Context, e *model.TxEvent) (*eh_tx.CallbackReq, error) {
	var event comm.Event
	if err := proto.Unmarshal([]byte(e.TxMsg), &event); err != nil {
		log.With(ctx).Errorf("unmarshal err:%v", err)
		return nil, err
	}
	return &eh_tx.CallbackReq{
		Event: &event,
		Tx: &comm.Tx{
			EventId: e.EventID,
			Status:  comm.TxStatus(e.TxStatus),
		},
	}, nil
}

// convertCallBackRsp is used to convert call back RPC response
func (s *Transaction) convertCallBackRsp(ctx context.Context, rsp *eh_tx.CallbackRsp, e *model.TxEvent) error {
	if rsp.Ret != nil && rsp.Ret.Code != 0 {
		return fmt.Errorf("callback err:%v", errcode.ConvertErr(rsp.Ret))
	}
	if rsp.TxStatus == comm.TxStatus_TX_STATUS_COMMIT {
		tx, err := s.Commit(ctx, &define.CommitReq{
			EventID: e.EventID,
			Trigger: rsp.Trigger,
		})
		if tx != nil && (tx.Status == TxStatusCommit || tx.Status == TxStatusRollback ||
			tx.Status == TxStatusSuccess) {
			return nil
		}
		return err
	}
	if rsp.TxStatus == comm.TxStatus_TX_STATUS_ROLLBACK {
		tx, err := s.Rollback(ctx, &define.RollbackReq{
			EventID: e.EventID,
		})
		if tx != nil && (tx.Status == TxStatusCommit || tx.Status == TxStatusRollback ||
			tx.Status == TxStatusSuccess) {
			return nil
		}
		return err
	}
	return errors.New("submit status err")
}

func (s *Transaction) verifyPrepare(req *define.PrepareReq) error {
	return nil
}

// makePrepareTxEvent is used to create a transaction event
func (s *Transaction) makePrepareTxEvent(req *define.PrepareReq) (*model.TxEvent, error) {
	eventID, err := s.handler.BizNO.NextID()
	if err != nil {
		return nil, err
	}
	e := &model.TxEvent{
		EventID:  eventID,
		AppID:    req.Event.AppId,
		TopicID:  req.Event.TopicId,
		TxStatus: int(TxStatusPrepare),
	}
	// initialize callback configuration
	if req.TxCallback != nil {
		e.CbCfType = int(CbCfTypeConfigParam)
		e.CbPtType = int(req.TxCallback.Endpoint.ProtocolType)
		e.CbAddr = req.TxCallback.Endpoint.Address
		e.CbTimeout = req.TxCallback.Endpoint.Timeout.AsDuration().Milliseconds()
		if req.TxCallback.Endpoint.Timeout.AsDuration().Milliseconds() == 0 {
			e.CbTimeout = DefaultCbTimeout
		}
		e.CbInterval = req.TxCallback.CallbackInterval.AsDuration().Milliseconds()
		if req.TxCallback.CallbackInterval.AsDuration().Milliseconds() == 0 {
			e.CbInterval = DefaultCbInterval
		}
		return e, nil
	}
	e.CbCfType = int(CbCfTypeConfigPlatform)

	producerConfInfo, ok := s.getProducerConf(req.Event.AppId, req.Event.TopicId)
	if !ok {
		return nil, fmt.Errorf("conf not found")
	}
	log.Debugf("producer info: %+v", producerConfInfo)
	e.CbPtType = producerConfInfo.TxProtocolType
	e.CbAddr = producerConfInfo.TxAddress
	e.CbTimeout = producerConfInfo.TxTimeout
	e.CbInterval = producerConfInfo.TxCallbackInterval
	return e, nil
}

// SendCallBackDelayMsg is used to send a delay message
func (s *Transaction) SendCallBackDelayMsg(msg *define.TxEventMsg) {
	ctx := msg.Ctx.(context.Context)
	b, err := s.makeEvhubMsg(ctx, msg)
	if err != nil {
		log.With(ctx).Errorf("marshal err:%v msg:%+v", msg)
		return
	}
	log.With(ctx).Debugf("tx event msg:%+v", msg)
	for i := 0; i < RetryCount; i++ {
		// send message
		if err = s.handler.DelayConnector.SendDelayMsg(ctx, topicutil.DelayTopic(msg.AppID,
			msg.TopicID), b, time.Duration(msg.DelayTimeMs)*time.Millisecond); err != nil {
			log.With(ctx).Errorf("send msg err:%v msg:%+v", err, msg)
			continue
		}
		break
	}
	if err != nil {
		log.With(ctx).Errorf("send msg err err:%v", err)
	}
}

// makeEvhubMsg is used to create event messages
func (s *Transaction) makeEvhubMsg(ctx context.Context, txEventMsg *define.TxEventMsg) ([]byte, error) {
	msg := &comm.EvhubMsg{
		EventId: txEventMsg.EventID,
		Status:  comm.EventStatus_EVENT_STATUS_TX_CALLBACK_IN_DELAY_QUEUE,
		Event: &comm.Event{
			AppId:   txEventMsg.AppID,
			TopicId: txEventMsg.TopicID,
			EventComm: &comm.EventComm{
				EventSource: &comm.EventSource{
					CreateTime: timestamppb.New(time.Now()),
				},
			},
		},
		DispatcherInfo: &comm.EventDispatcherInfo{
			RetryTimes: txEventMsg.RetryCount,
		},
	}
	// marshal event
	msgByte, err := proto.Marshal(msg)
	if err != nil {
		log.With(ctx).Infof("encode err:%v", err)
		return nil, err
	}
	return msgByte, nil
}

// AsyncReportEvent is used to report events
func (s *Transaction) AsyncReportEvent(args interface{}) {
	msg := args.(*define.TxEventMsg)
	ctx := msg.Ctx.(context.Context)
	log.With(ctx).Debugf("%+v", msg)

	if msg.Type == AsyncCallBack {
		s.SendCallBackDelayMsg(msg)
		return
	}

	e := &model.TxEvent{
		EventID: msg.EventID,
	}
	var result *gorm.DB
	for i := 0; i < RetryCount; i++ {
		engine := s.db.Engine
		result = engine.First(e)
		if result.Error != nil {
			log.With(ctx).Errorf("get tx event err:%+v", result.Error)
			continue
		}
		break
	}
	if result == nil || result.RowsAffected == 0 {
		log.With(ctx).Errorf("not get tx event e:%v", e)
		return
	}
	log.With(ctx).Debugf("first tx event e:%+v", e)
	s.commitEvent(ctx, e, msg)
}

// commitEvent is used to commit a transaction event
func (s *Transaction) commitEvent(ctx context.Context, e *model.TxEvent, txEventMsg *define.TxEventMsg) {
	var event comm.Event
	if err := proto.Unmarshal([]byte(e.TxMsg), &event); err != nil {
		log.With(ctx).Errorf("unmarshal err:%v", err)
		return
	}
	req := &define.ReportReq{
		Event:   &event,
		Trigger: txEventMsg.Trigger,
	}
	var err error
	// transaction commit
	err = s.db.Engine.Transaction(func(tx *gorm.DB) error {
		var result *gorm.DB
		for i := 0; i < RetryCount; i++ {
			result = s.db.Engine.Table(e.TableName()).Where("event_id = ? and tx_status = ?",
				e.EventID, TxStatusCommit).Update("tx_status", TxStatusSuccess)
			if result.Error != nil {
				log.With(ctx).Errorf("update tx_status success err:%+v req:%+v", result.Error, req)
				continue
			}
			break
		}
		if result == nil || result.Error != nil {
			return fmt.Errorf("update tx_status success r:%+v req:%+v", result, req)
		}
		for i := 0; i < RetryCount; i++ {
			// handle events messages
			if err = s.handler.handleMsg(ctx, req, txEventMsg.EventID); err != nil {
				log.With(ctx).Errorf("handleMsg err iRet:%+v req:%+v", err, req)
				continue
			}
			break
		}
		if err != nil {
			return fmt.Errorf("handleMsg err err:%+v req:%+v", err, req)
		}
		return nil
	})

	if err != nil {
		log.With(ctx).Errorf("db tx exec err:%v req:%+v", err, req)
	}
}

// Prepare is used to prepare a transaction
func (s *Transaction) Prepare(ctx context.Context, req *define.PrepareReq) (string, error) {
	if err := s.verifyPrepare(req); err != nil {
		return "", codes.Errorf(errcode.CommParamInvalid, "verify:%v", err)
	}
	txEvent, err := s.makePrepareTxEvent(req)
	if err != nil {
		log.With(ctx).Errorf("make prepare tx event err:%v", err)
		return "", codes.Errorf(errcode.CommParamInvalid, "verify:%v", err)
	}
	txMsg, err := proto.Marshal(req.Event)
	if err != nil {
		return txEvent.EventID, codes.Errorf(errcode.CommParamInvalid, "verify:%v", err)
	}
	txEvent.TxMsg = string(txMsg)
	result := s.db.Engine.Create(txEvent)
	if result.Error != nil {
		log.With(ctx).Infof("create event failed: %+v req:%+v", result.Error, req)
		return txEvent.EventID, codes.Errorf(errcode.TxProduceFail, "produce failed")
	}

	if err := s.routinePool.Invoke(&define.TxEventMsg{
		Ctx:         ctx,
		AppID:       req.Event.AppId,
		TopicID:     req.Event.TopicId,
		EventID:     txEvent.EventID,
		DelayTimeMs: uint32(txEvent.CbInterval),
		Type:        AsyncCallBack,
	}); err != nil {
		log.With(ctx).Errorf("asyncReport error:%v", err)
	}
	return txEvent.EventID, nil
}

// verifyCommit is used to verify commit request
func (s *Transaction) verifyCommit(req *define.CommitReq) error {
	if req == nil {
		return fmt.Errorf("req is empty")
	}
	if req.EventID == "" {
		return fmt.Errorf("eventId is empty")
	}
	if req.Trigger == nil {
		return fmt.Errorf("trigger is empty")
	}
	return nil
}

// updateTxStatus is used to update the status of a transaction message to the server
func (s *Transaction) updateTxStatus(ctx context.Context, txEvent *model.TxEvent) (*comm.Tx, error) {
	r := s.db.Engine.Table(txEvent.TableName()).Where("event_id = ? and tx_status = ?",
		txEvent.EventID, int(TxStatusPrepare)).Updates(txEvent)
	if r.Error != nil {
		log.With(ctx).Errorf("update tx event err:%+v", r.Error)
		return nil, codes.Errorf(errcode.SystemErr, "update status err")
	}
	if r.RowsAffected != 0 {
		return nil, nil
	}
	e := &model.TxEvent{
		EventID: txEvent.EventID,
	}
	result := s.db.Engine.First(e)
	if result.Error != nil {
		return nil, codes.Errorf(errcode.SystemErr, "find tx msg err")
	}
	if result.RowsAffected == 0 {
		return nil, codes.Errorf(errcode.TxUnRegistered, "tx unregistered")
	}
	if txEvent.TxStatus == int(TxStatusCommit) {
		return &comm.Tx{
			EventId: txEvent.EventID,
			Status:  comm.TxStatus(e.TxStatus),
		}, codes.Errorf(errcode.TxCommittedFail, "tx commit failed")
	}
	return &comm.Tx{
		EventId: txEvent.EventID,
		Status:  comm.TxStatus(e.TxStatus),
	}, codes.Errorf(errcode.TxRolledBackFail, "tx rollback failed")
}

// Commit is used to commit a transaction
func (s *Transaction) Commit(ctx context.Context, req *define.CommitReq) (*comm.Tx, error) {
	if err := s.verifyCommit(req); err != nil {
		return nil, codes.Errorf(errcode.CommParamInvalid, "verify:%v", err)
	}
	trigger, err := proto.Marshal(req.Trigger)
	if err != nil {
		return nil, codes.Errorf(errcode.CommParamInvalid, "verify:%v", err)
	}
	txEvent := &model.TxEvent{
		EventID:   req.EventID,
		TxStatus:  int(TxStatusCommit),
		TxTrigger: string(trigger),
	}

	tx, err := s.updateTxStatus(ctx, txEvent)
	if err != nil {
		return tx, err
	}
	if err := s.routinePool.Invoke(&define.TxEventMsg{
		Ctx:     ctx,
		EventID: req.EventID,
		Type:    AsyncReportEvent,
		Trigger: req.Trigger,
	}); err != nil {
		log.With(ctx).Errorf("asyncReport error:%v", err)
	}

	return &comm.Tx{
		EventId: req.EventID,
		Status:  TxStatusCommit,
	}, nil
}

// verifyRollback is used to verify that the transaction is rolled back
func (s *Transaction) verifyRollback(req *define.RollbackReq) error {
	if req.EventID == "" {
		return fmt.Errorf("event id is required")
	}
	return nil
}

// Rollback is used rollback the transaction
func (s *Transaction) Rollback(ctx context.Context, req *define.RollbackReq) (*comm.Tx, error) {
	if err := s.verifyRollback(req); err != nil {
		return nil, codes.Errorf(errcode.CommParamInvalid, "verify:%v", err)
	}
	txEvent := &model.TxEvent{
		EventID:  req.EventID,
		TxStatus: int(TxStatusRollback),
	}
	tx, err := s.updateTxStatus(ctx, txEvent)
	if err != nil {
		return tx, err
	}
	return &comm.Tx{
		EventId: req.EventID,
		Status:  comm.TxStatus_TX_STATUS_ROLLBACK,
	}, nil
}

// periodCallBack is used to check whether it is final
func (s *Transaction) periodCallBack() {
	s.cbHeartbeat = time.NewTicker(time.Duration(CbHeartbeatTimeOut) * time.Second)
	for range s.cbHeartbeat.C {
		txEvent := &model.TxEvent{}
		var txEventList []*model.TxEvent
		engine := s.readDB.Engine
		r := engine.Table(txEvent.TableName()).
			Where("tx_status = ? and UNIX_TIMESTAMP(sys_utime) < UNIX_TIMESTAMP(?)", int(TxStatusPrepare),
				time.Now().Add(time.Second*CbExpire).Format(TimeStrFormat)).Find(&txEventList).Limit(PeriodLimit)
		if r.Error != nil {
			log.Errorf("period call back err:%v", r.Error)
			continue
		}
		for _, txEvent := range txEventList {
			if txEvent.SysCtime.Add(time.Duration(txEvent.CbTimeout) *
				time.Millisecond).After(time.Now()) {
				continue
			}
			ctx := log.SetDye(context.Background(), "", "", txEvent.EventID, 0, 0)
			log.With(ctx).Infof("period call back event:%+v", txEvent)
			if err := s.cbCgTypeConfigParam(ctx, txEvent); err != nil {
				txEventMsg := &define.TxEventMsg{
					Ctx:         ctx,
					EventID:     txEvent.EventID,
					AppID:       txEvent.AppID,
					TopicID:     txEvent.TopicID,
					DelayTimeMs: uint32(txEvent.CbInterval),
					RetryCount:  uint32(txEvent.RetryCount),
				}
				s.SendCallBackDelayMsg(txEventMsg)
				log.Infof("call back event msg:%+v err:%v", txEvent, err)
			}
		}
	}
}

// periodCommitFail is used to check whether it is final
func (s *Transaction) periodCommitFail() {
	s.cfHeartbeat = time.NewTicker(time.Duration(CfHeartbeatTimeOut) * time.Second)
	for range s.cfHeartbeat.C {
		txEvent := &model.TxEvent{}
		var txEventList []*model.TxEvent
		engine := s.readDB.Engine
		r := engine.Table(txEvent.TableName()).
			Where("tx_status = ? and UNIX_TIMESTAMP(sys_utime) < UNIX_TIMESTAMP(?)", int(TxStatusCommit),
				time.Now().Add(time.Second*CbExpire).Format(TimeStrFormat)).Find(&txEventList).Limit(PeriodLimit)
		if r.Error != nil {
			log.Errorf("period call back err:%v", r.Error)
			continue
		}
		for _, txEvent := range txEventList {
			log.Infof("period commit tx event:%+v", txEvent)
			var trigger comm.EventTrigger
			if err := proto.Unmarshal([]byte(txEvent.TxTrigger), &trigger); err != nil {
				log.Errorf("unmarshal txEvent:%+v err:%v", txEvent, err)
				continue
			}
			s.commitEvent(context.Background(), txEvent, &define.TxEventMsg{
				EventID: txEvent.EventID,
				Trigger: &trigger,
			})
		}
	}
}

// CallBack is used to call back
func (s *Transaction) CallBack(ctx context.Context, req interface{}) (interface{}, error) {
	in, ok := req.(*define.CallbackReq)
	if !ok {
		return nil, fmt.Errorf("convert err")
	}

	if err := s.CallBackInternal(ctx, in.TxEventMsg); err != nil {
		return &define.CallbackRsp{}, err
	}
	return &define.CallbackRsp{}, nil
}

// CallBackInternal
func (s *Transaction) CallBackInternal(ctx context.Context, msg *define.TxEventMsg) error {
	e := &model.TxEvent{
		EventID: msg.EventID,
	}
	result := s.db.Engine.First(e)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		log.With(ctx).Errorf("not get tx event msg:%+v", msg)
		return nil
	}
	if comm.TxStatus(e.TxStatus) != TxStatusPrepare {
		return nil
	}
	if err := s.cbCgTypeConfigParam(ctx, e); err != nil {
		log.With(ctx).Infof("callback err:%v", err)
		return err
	}
	return nil
}

// cbCgTypeConfigParam
func (s *Transaction) cbCgTypeConfigParam(ctx context.Context, e *model.TxEvent) error {
	var err error
	packProtocol := define.PackProtocolGrpc
	switch comm.ProtocolType(e.CbPtType) {
	case comm.ProtocolType_PROTOCOL_TYPE_GRPC:
		err = s.callbackGrpc(ctx, e)
		packProtocol = define.PackProtocolGrpc
	case comm.ProtocolType_PROTOCOL_TYPE_HTTP_JSON:
		err = s.callbackHTTPJSON(ctx, e)
		packProtocol = define.PackProtocolHTTPJSON
	default:
		err = s.callbackGrpc(ctx, e)
	}
	if err != nil {
		r := s.db.Engine.Table(e.TableName()).Where("event_id = ?",
			e.EventID).Update("retry_count", e.RetryCount+1)
		if r.Error != nil {
			log.With(ctx).Errorf("update tx event err:%+v packProtocol:%v", r.Error, packProtocol)
			return r.Error
		}
	}
	return err
}

// callbackGrpc is used to grpc callback
func (s *Transaction) callbackGrpc(ctx context.Context, e *model.TxEvent) error {
	addr := addrutil.ParseAddr(e.CbAddr)
	conn, err := grpc.Dial(&grpc.ClientConfig{Addr: addr, Timeout: time.Duration(e.CbTimeout) * time.Millisecond})
	if err != nil {
		return err
	}
	defer conn.Close()
	req, err := s.makeCallbackReq(ctx, e)
	if err != nil {
		return err
	}
	rsp, err := eh_tx.NewEvhubTransactionClient(conn).Callback(ctx, req)
	if err != nil {
		return err
	}
	return s.convertCallBackRsp(ctx, rsp, e)
}

// callbackHTTPJSON is used to callback
func (s *Transaction) callbackHTTPJSON(ctx context.Context, e *model.TxEvent) error {
	return s.callbackHTTP(ctx, e, comm.ProtocolType_PROTOCOL_TYPE_HTTP_JSON)
}

//  callbackHTTP is used to send a request
func (s *Transaction) callbackHTTP(ctx context.Context, e *model.TxEvent, protocolType comm.ProtocolType) error {
	addr := addrutil.ParseAddr(e.CbAddr)
	if strings.HasPrefix(e.CbAddr, "dns://") ||
		strings.HasPrefix(e.CbAddr, "ip://") {
		addr = "http://" + addr
	}
	req, err := s.makeCallbackReq(ctx, e)
	if err != nil {
		return err
	}
	reqBody, err := s.makeHTTPReq(req, protocolType)
	if err != nil {
		return err
	}
	client := http.Client{
		Timeout: time.Millisecond * time.Duration(e.CbTimeout),
	}
	httpRsp, err := client.Post(addr, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	defer httpRsp.Body.Close()
	rspBody, err := ioutil.ReadAll(httpRsp.Body)
	if err != nil {
		return err
	}
	rsp, err := s.makeHTTPRsp(rspBody, protocolType)
	if err != nil {
		return err
	}
	return s.convertCallBackRsp(ctx, rsp, e)
}

// makeHTTPRsp returns an HTTP response
func (s *Transaction) makeHTTPRsp(rspBody []byte, protocolType comm.ProtocolType) (*eh_tx.CallbackRsp, error) {
	rsp := &eh_tx.CallbackRsp{}
	if protocolType == comm.ProtocolType_PROTOCOL_TYPE_HTTP_PB {
		if err := proto.Unmarshal(rspBody, rsp); err != nil {
			return nil, err
		}
		return rsp, nil
	}
	if protocolType == comm.ProtocolType_PROTOCOL_TYPE_HTTP_JSON {
		if err := json.Unmarshal(rspBody, rsp); err != nil {
			return nil, err
		}
		return rsp, nil
	}
	return nil, fmt.Errorf("protocol type not supported")
}

// makeHTTPReq returns an HTTP request
func (s *Transaction) makeHTTPReq(req *eh_tx.CallbackReq, protocolType comm.ProtocolType) ([]byte, error) {
	if protocolType == comm.ProtocolType_PROTOCOL_TYPE_HTTP_PB {
		reqBody, err := proto.Marshal(req)
		if err != nil {
			return nil, err
		}
		return reqBody, nil
	}
	if protocolType == comm.ProtocolType_PROTOCOL_TYPE_HTTP_JSON {
		reqBody, err := json.Marshal(req)
		if err != nil {
			return nil, err
		}
		return reqBody, nil
	}
	return nil, fmt.Errorf("protocol type not supported")
}

// Stop stops the transaction gracefully.
func (s *Transaction) Stop() error {
	s.cbHeartbeat.Stop()
	s.cfHeartbeat.Stop()
	return nil
}
