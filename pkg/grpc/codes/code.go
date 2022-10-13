package codes

import (
	"fmt"
	"strconv"
	"sync/atomic"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	// messages stores the messages of the codes, structured as a map[int]string.
	messages atomic.Value
)

// Register registers the code message.
func Register(m map[int]string) {
	messages.Store(m)
}

// A Code is an int error code.
type Code int

func (c Code) Error() string {
	return strconv.FormatInt(int64(c), 10)
}

// Code returns the error code.
func (c Code) Code() int {
	return int(c)
}

// Message returns the error message.
func (c Code) Message() string {
	if m, ok := messages.Load().(map[int]string); ok {
		if msg, ok := m[c.Code()]; ok {
			return msg
		}
	}
	return c.Error()
}

// Details returns the details.
func (c Code) Details() []interface{} {
	return nil
}

// GRPCStatus returns the Status.
func (c Code) GRPCStatus() *status.Status {
	return status.New(codes.Code(c), c.Message())
}

// Error returns an error representing c and msg.  If c is OK, returns nil.
func Error(c Code, msg string) error {
	return status.New(codes.Code(c), msg).Err()
}

// Errorf returns Error(c, fmt.Sprintf(format, a...)).
func Errorf(c Code, format string, a ...interface{}) error {
	return Error(c, fmt.Sprintf(format, a...))
}

// EqualError equal error.
func EqualError(code Code, err error) bool {
	if err == nil {
		return code.Code() == int(codes.OK)
	}
	return code.Code() == int(status.Convert(err).Code())
}
