package runtime

import (
	corelog "cosmossdk.io/core/log"
	"cosmossdk.io/log"
)

type LoggerService struct {
	logger log.Logger
}

func (ls LoggerService) Info(msg string, keyVals ...any) {
	ls.logger.Info(msg, keyVals...)
}

func (ls LoggerService) Error(msg string, keyVals ...any) {
	ls.logger.Error(msg, keyVals...)
}

func (ls LoggerService) Debug(msg string, keyVals ...any) {
	ls.logger.Debug(msg, keyVals...)
}

func (ls LoggerService) With(keyVals ...any) corelog.Service {
	return LoggerService{logger: ls.logger.With(keyVals...)}
}

func (ls LoggerService) Impl() any {
	return ls.logger
}
