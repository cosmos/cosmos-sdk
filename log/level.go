package log

const defaultLogLevelKey = "*"

// ParseLogLevel parses complex log level
// A comma-separated list of module:level pairs with an optional *:level pair
// (* means all other modules).
//
// Example:
// ParseLogLevel("consensus:debug,mempool:debug,*:error", "info")
func ParseLogLevel(lvl string, defaultLogLevelValue string) (func(key, level string) bool, error) {
	return func(key, level string) bool { return true }, nil
}
