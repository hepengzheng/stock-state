package logutil

import (
	"runtime/debug"

	"go.uber.org/zap"

	"github.com/hepengzheng/stock-state/pkg/logger"
)

func LogPanic() {
	if r := recover(); r != nil {
		logger.Error("panic occurs",
			zap.Any("err", r),
			zap.String("stack", string(debug.Stack())),
		)
	}
}
