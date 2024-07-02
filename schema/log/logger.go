package log

type Logger interface {
	// Info takes a message and a set of key/value pairs and logs with level INFO.
	// The key of the tuple must be a string.
	Info(msg string, keyVals ...interface{})

	// Warn takes a message and a set of key/value pairs and logs with level WARN.
	// The key of the tuple must be a string.
	Warn(msg string, keyVals ...interface{})

	// Error takes a message and a set of key/value pairs and logs with level ERR.
	// The key of the tuple must be a string.
	Error(msg string, keyVals ...interface{})

	// Debug takes a message and a set of key/value pairs and logs with level DEBUG.
	// The key of the tuple must be a string.
	Debug(msg string, keyVals ...interface{})
}

type NoopLogger struct{}

func (n NoopLogger) Info(msg string, keyVals ...interface{}) {}

func (n NoopLogger) Warn(msg string, keyVals ...interface{}) {}

func (n NoopLogger) Error(msg string, keyVals ...interface{}) {}

func (n NoopLogger) Debug(msg string, keyVals ...interface{}) {}

var _ Logger = NoopLogger{}
