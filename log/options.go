package log

var defaultConfig = Config{
	Filter:     nil,
	OutputJSON: false,
}

// LoggerConfig defines configuration for the logger.
type Config struct {
	Filter     FilterFunc
	OutputJSON bool
}

type Option func(*Config)

// FilterOption sets the filter for the Logger.
func FilterOption(filter FilterFunc) Option {
	return func(cfg *Config) {
		cfg.Filter = filter
	}
}

// OutputJSONOption sets the output of the logger to JSON.
func OutputJSONOption() Option {
	return func(cfg *Config) {
		cfg.OutputJSON = true
	}
}
