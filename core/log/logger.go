package log

const ModuleKey = "module"

// Logger is the Cosmos SDK logger interface.
// It maintains as much backward compatibility with the CometBFT logger as possible.
// cosmossdk.io/log is the implementation provided by the Cosmos SDK
// All functionalities of the logger are available through the Impl() method.
type Logger interface {
	LoggerBase

	// With returns a new wrapped logger with additional context provided by a set.
	With(keyVals ...any) Logger
}

// LoggerBase defines basic logger functionality.
type LoggerBase interface {
	// Info takes a message and a set of key/value pairs and logs with level INFO.
	// The key of the tuple must be a string.
	Info(msg string, keyVals ...any)

	// Warn takes a message and a set of key/value pairs and logs with level WARN.
	// The key of the tuple must be a string.
	Warn(msg string, keyVals ...any)

	// Error takes a message and a set of key/value pairs and logs with level ERR.
	// The key of the tuple must be a string.
	Error(msg string, keyVals ...any)

	// Debug takes a message and a set of key/value pairs and logs with level DEBUG.
	// The key of the tuple must be a string.
	Debug(msg string, keyVals ...any)

	// Impl returns the underlying logger implementation.
	// It is used to access the full functionalities of the underlying logger.
	// Advanced users can type cast the returned value to the actual logger.
	Impl() any
}

// LoggerV2 is the newer Cosmos SDK logger interface. It can be used in libraries without any need to import
// the core or log packages directly.
type LoggerV2 interface {
	LoggerBase

	// WithV2 returns a type which can be cast to LoggerV2 with additional context provided by a set.
	// any is returned instead of LoggerV2 to avoid the need for log users to import the log package directly.
	WithV2(keyVals ...any) any
}
