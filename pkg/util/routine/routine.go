package routine

import (
	"runtime"

	"github.com/tencentmusic/evhub/pkg/log"
)

// GoWithRecovery go recover panic
func GoWithRecovery(f func()) {
	go func() {
		defer func() {
			if e := recover(); e != nil {
				// stack
				stack := make([]byte, 16*1024*1024)
				stack = stack[:runtime.Stack(stack, false)]
				log.Errorf("recover stack: %s, err: %s", string(stack), e)
			}
		}()
		//  function
		f()
	}()
}

// Recover is used to recover
func Recover() {
	if e := recover(); e != nil {
		stack := make([]byte, 16*1024*1024)
		stack = stack[:runtime.Stack(stack, false)]
		log.Errorf("recover stack: %s, err: %s", string(stack), e)
	}
}
