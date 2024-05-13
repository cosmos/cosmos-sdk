package log

const defaultLogLevelKey = "*"

// FilterFunc is a function that returns true if the log level is filtered for the given key
// When the filter returns true, the log entry is discarded.
type FilterFunc func(key, level string) bool

// ParseLogLevel parses complex log level
// A comma-separated list of module:level pairs with an optional *:level pair
// (* means all other modules).
//
// Example:
// ParseLogLevel("consensus:debug,mempool:debug,*:error")
//
// This function attempts to keep the same behavior as the CometBFT ParseLogLevel
// However the level `none` is replaced by `disabled`.
func ParseLogLevel(levelStr string) (FilterFunc, error) {
	panic("impl")
}
