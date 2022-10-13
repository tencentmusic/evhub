package grpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClientDial(t *testing.T) {
	config := &ClientConfig{Addr: ":0"}
	_, err := Dial(config)
	assert.NoError(t, err)
}

func TestDialEmptyConfig(t *testing.T) {
	_, err := Dial(nil)
	assert.Error(t, err)
}

func TestDialEmptyAddr(t *testing.T) {
	_, err := Dial(&ClientConfig{})
	assert.Error(t, err)
}
