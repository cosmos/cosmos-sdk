package log

import "github.com/rs/zerolog"

// defaultConfig has all the options disabled.
var defaultConfig = Config{
	Level:      zerolog.NoLevel,
	Filter:     nil,
	OutputJSON: false,
}

// LoggerConfig defines configuration for the logger.
type Config struct {
	Level      zerolog.Level
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

// LevelOption sets the level for the Logger.
func LevelOption(level zerolog.Level) Option {
	return func(cfg *Config) {
		cfg.Level = level
	}
}

// OutputJSONOption sets the output of the logger to JSON.
func OutputJSONOption() Option {
	return func(cfg *Config) {
		cfg.OutputJSON = true
	}
}
