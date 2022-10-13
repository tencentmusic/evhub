package log

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLevelUnmarshalText(t *testing.T) {
	tests := []struct {
		text    []byte
		wantErr bool
	}{
		{[]byte(""), true},
		{[]byte("bad"), true},
		{[]byte("debug"), false},
		{[]byte("info"), false},
		{[]byte("warn"), false},
		{[]byte("error"), false},
		{[]byte("panic"), false},
		{[]byte("fatal"), false},
	}
	var level Level
	for _, tt := range tests {
		assert.Equal(t, level.UnmarshalText(tt.text) != nil, tt.wantErr)
	}
}
