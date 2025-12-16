package log

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
)

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
// This function attempts to keep the same behavior as the CometBFT ParseLogLevel.
func ParseLogLevel(levelStr string) (FilterFunc, error) {
	if levelStr == "" {
		return nil, errors.New("empty log level")
	}

	// prefix simple one word levels (e.g. "info") with "*"
	l := levelStr
	if !strings.Contains(l, ":") {
		l = defaultLogLevelKey + ":" + l
	}

	// parse and validate the levels
	filterMap := make(map[string]slog.Level)
	list := strings.Split(l, ",")
	for _, item := range list {
		moduleAndLevel := strings.Split(item, ":")
		if len(moduleAndLevel) != 2 {
			return nil, fmt.Errorf("expected list in a form of \"module:level\" pairs, given pair %s, list %s", item, list)
		}

		module := moduleAndLevel[0]
		levelName := moduleAndLevel[1]

		if _, ok := filterMap[module]; ok {
			return nil, fmt.Errorf("duplicate module %s in log level list %s", module, list)
		}

		level, err := ParseLevel(levelName)
		if err != nil {
			return nil, fmt.Errorf("invalid log level %s in log level list %s", levelName, list)
		}

		filterMap[module] = level
	}

	// If there's only a default level and no module-specific levels, no filter needed
	if len(filterMap) == 1 {
		if _, ok := filterMap[defaultLogLevelKey]; ok {
			return nil, nil
		}
	}

	filterFunc := func(key, lvl string) bool {
		level, ok := filterMap[key]
		if !ok {
			// no level filter for this key, check if there is a default level filter
			level, ok = filterMap[defaultLogLevelKey]
			if !ok {
				return false
			}
		}

		msgLevel, err := ParseLevel(lvl)
		if err != nil {
			// If we can't parse the level, don't filter
			return false
		}

		return msgLevel < level
	}

	return filterFunc, nil
}

// ParseLevel parses a level string into a slog.Level.
// Supported levels: debug, info, warn, error, disabled/none.
func ParseLevel(s string) (slog.Level, error) {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error", "err":
		return slog.LevelError, nil
	case "disabled", "none":
		// Use a very high level to effectively disable logging
		return slog.Level(100), nil
	default:
		return slog.LevelInfo, fmt.Errorf("unknown log level: %s, valid levels are: debug, info, warn, error, disabled", s)
	}
}

// LevelToString converts a slog.Level to its string representation.
func LevelToString(level slog.Level) string {
	switch {
	case level <= slog.LevelDebug:
		return "debug"
	case level <= slog.LevelInfo:
		return "info"
	case level <= slog.LevelWarn:
		return "warn"
	case level <= slog.LevelError:
		return "error"
	default:
		return "disabled"
	}
}
