package errcode

import (
	"github.com/tencentmusic/evhub/pkg/gen/proto/comm"
	"github.com/tencentmusic/evhub/pkg/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	// CommParamInvalid is returned when a parameter is invalid
	CommParamInvalid = 10000
	// CheckIdempotentFail is returned when idempotent check fails
	CheckIdempotentFail = 10001
	// ProduceFail is returned when producer fails
	ProduceFail = 20000
	// DelayProduceFail is returned when producer fails
	DelayProduceFail = 20001
	// TxProduceFail is returned when tx producer fails
	TxProduceFail = 21000
	// TxCommittedFail is returned when tx commit fails
	TxCommittedFail = 21001
	// TxRolledBackFail is returned when tx rollback fails
	TxRolledBackFail = 21002
	// TxUnRegistered is returned when tx unregistered
	TxUnRegistered = 21002
	// EventCancel is returned when event cancel
	EventCancel = 30000
	// SystemErr is returned when system error
	SystemErr = 20005
)

var codeName = map[int32]string{
	CommParamInvalid:    "CommParamInvalid",
	CheckIdempotentFail: "CheckIdempotentFail",
	ProduceFail:         "ProduceFail",
	DelayProduceFail:    "DelayProduceFail",
	TxProduceFail:       "TxProduceFail",
	TxCommittedFail:     "TxCommittedFail",
	TxRolledBackFail:    "TxRolledBackFail",
	EventCancel:         "EventCancel",
	SystemErr:           "SystemErr",
}

// CodeError returns the error associated with the code
func CodeError(code int32) error {
	return codes.Error(codes.Code(code), codeName[code])
}

// ConvertRet returns the result of converting error
func ConvertRet(err error) *comm.Result {
	return &comm.Result{
		Code: int32(status.Convert(err).Code()),
		Msg:  status.Convert(err).Message(),
	}
}

// ConvertErr returns the error of converting result
func ConvertErr(ret *comm.Result) error {
	return codes.Error(codes.Code(ret.Code), ret.Msg)
}
