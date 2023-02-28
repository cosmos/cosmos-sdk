package log

import "errors"

// FilterFunc is a function that returns true if the log level is filtered for the given key
// When the filter returns true, the log entry is discarded
type FilterFunc func(key, level string) bool

const defaultLogLevelKey = "*"

// ParseLogLevel parses complex log level
// A comma-separated list of module:level pairs with an optional *:level pair
// (* means all other modules).
//
// Example:
// ParseLogLevel("consensus:debug,mempool:debug,*:error", "info")
//
// This function attemps to keep the same behavior as the CometBFT ParseLogLevel
func ParseLogLevel(lvl string, defaultLogLevelValue string) (FilterFunc, error) {
	if lvl == "" {
		return nil, errors.New("empty log level")
	}

	return func(key, level string) bool { return true }, nil
}
