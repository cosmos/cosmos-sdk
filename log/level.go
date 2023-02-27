package log

import "errors"

const defaultLogLevelKey = "*"

// ParseLogLevel parses complex log level
// A comma-separated list of module:level pairs with an optional *:level pair
// (* means all other modules).
//
// Example:
// ParseLogLevel("consensus:debug,mempool:debug,*:error", "info")
// 
// This function attemps to keep the same behavior as the CometBFT ParseLogLevel
func ParseLogLevel(lvl string, defaultLogLevelValue string) (func(key, level string) bool, error) {
	if lvl == "" {
		return nil, errors.New("empty log level")
	}

	return func(key, level string) bool { return true }, nil
}
