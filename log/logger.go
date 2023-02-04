package log

import cmlog "github.com/cometbft/cometbft/libs/log"

// Logger is the interface that wraps the basic logging methods.
type Logger interface {
	Debug(msg string, keyvals ...interface{})
	Info(msg string, keyvals ...interface{})
	Error(msg string, keyvals ...interface{})

	With(keyvals ...interface{}) cmlog.Logger
}
