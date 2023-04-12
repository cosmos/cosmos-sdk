package log

type Service interface {
	Logger() Logger
}

// Logger is the Cosmos SDK logger interface (copied from cosmossdk.io/log).
// It maintains as much backward compatibility with the CometBFT logger as possible.
// All functionalities of the logger are available through the Impl() method.
type Logger interface {
	// Info takes a message and a set of key/value pairs and logs with level INFO.
	// The key of the tuple must be a string.
	Info(msg string, keyVals ...any)

	// Error takes a message and a set of key/value pairs and logs with level ERR.
	// The key of the tuple must be a string.
	Error(msg string, keyVals ...any)

	// Debug takes a message and a set of key/value pairs and logs with level DEBUG.
	// The key of the tuple must be a string.
	Debug(msg string, keyVals ...any)

	// With returns a new wrapped logger with additional context provided by a set.
	With(keyVals ...any) Logger

	// Impl returns the underlying logger implementation.
	// It is used to access the full functionalities of the underlying logger.
	// Advanced users can type cast the returned value to the actual logger.
	Impl() any
}
