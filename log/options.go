package log

var defaultLoggerConfig = LoggerConfig{
	Filter:     nil,
	OutputJSON: false,
}

// LoggerConfig defines configuration for the logger
type LoggerConfig struct {
	Filter     FilterFunc
	OutputJSON bool
}

type LoggerOption func(*LoggerConfig)

// FilterLoggerOption sets the filter for the Logger
func FilterLoggerOption(filter FilterFunc) LoggerOption {
	return func(cfg *LoggerConfig) {
		cfg.Filter = filter
	}
}

// OutputJSONLoggerOption sets the output of the Logger to JSON
func OutputJSONLoggerOption() LoggerOption {
	return func(cfg *LoggerConfig) {
		cfg.OutputJSON = true
	}
}
