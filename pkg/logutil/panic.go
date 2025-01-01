package logutil

import (
	"fmt"
	"runtime/debug"
)

func LogPanic() {
	if r := recover(); r != nil {
		fmt.Printf("panic occurs, err:%v, stack:%s\n", r, string(debug.Stack()))
	}
}
