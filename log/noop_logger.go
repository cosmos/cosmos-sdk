package log

import cmlog "github.com/cometbft/cometbft/libs/log"

var _ Logger = &NoOp{}

type NoOp struct{}

func NewNoOpLogger() *NoOp {
	return &NoOp{}
}

func (l NoOp) Debug(msg string, keyvals ...interface{}) {}
func (l NoOp) Info(msg string, keyvals ...interface{})  {}
func (l NoOp) Error(msg string, keyvals ...interface{}) {}

func (l NoOp) With(i ...interface{}) cmlog.Logger {
	return l
}
