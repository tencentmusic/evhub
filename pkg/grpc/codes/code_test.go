package codes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCodeError(t *testing.T) {
	tests := []struct {
		c    Code
		want string
	}{
		{Code(0), "0"},
		{Code(1), "1"},
		{Code(2), "2"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, tt.c.Error())
	}
}

func TestCodeMessage(t *testing.T) {
	m := map[int]string{0: "Code(0)", 1: "Code(1)", 2: "Code(2)"}
	Register(m)

	tests := []struct {
		c    Code
		want string
	}{
		{Code(0), m[0]},
		{Code(1), m[1]},
		{Code(2), m[2]},
		{Code(3), "3"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, tt.c.Message())
	}
}

func TestCodeDetails(t *testing.T) {
	code := Code(0)
	assert.Equal(t, []interface{}(nil), code.Details())
}

func TestEqualError(t *testing.T) {
	tests := []struct {
		code Code
		err  error
		want bool
	}{
		{Code(0), nil, true},
		{Code(1), Error(Code(1), "canceled"), true},
		{Code(2), Errorf(Code(2), "unknown %s", "test"), true},
		{Code(3), status.Error(codes.Code(3), "invalid argument"), true},
		{Code(4), Error(Code(5), ""), false},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, EqualError(tt.code, tt.err))
	}
}

func TestCodeGRPCStatus(t *testing.T) {
	Register(nil)
	tests := []struct {
		c    Code
		want *status.Status
	}{
		{Code(0), status.New(codes.Code(0), "0")},
		{Code(1), status.New(codes.Code(1), "1")},
		{Code(2), status.New(codes.Code(2), "2")},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, tt.c.GRPCStatus())
	}
}
