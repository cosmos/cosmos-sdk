package serverv2

import (
	"io"

	"github.com/rs/zerolog"

	"cosmossdk.io/core/server"
	"cosmossdk.io/log"
)

// NewLogger creates the default SDK logger.
// It reads the log level and format from the server context.
func NewLogger(cfg server.ConfigMap, out io.Writer) (log.Logger, error) {
	var opts []log.Option
	var (
		format  string
		noColor bool
		trace   bool
		level   string
	)
	if v, ok := cfg[FlagLogFormat]; ok {
		format = v.(string)
	}
	if v, ok := cfg[FlagLogNoColor]; ok {
		noColor = v.(bool)
	}
	if v, ok := cfg[FlagTrace]; ok {
		trace = v.(bool)
	}
	if v, ok := cfg[FlagLogLevel]; ok {
		level = v.(string)
	}

	if format == OutputFormatJSON {
		opts = append(opts, log.OutputJSONOption())
	}
	opts = append(opts,
		log.ColorOption(!noColor),
		log.TraceOption(trace),
	)

	// check and set filter level or keys for the logger if any
	if level == "" {
		return log.NewLogger(out, opts...), nil
	}

	logLvl, err := zerolog.ParseLevel(level)
	switch {
	case err != nil:
		// If the log level is not a valid zerolog level, then we try to parse it as a key filter.
		filterFunc, err := log.ParseLogLevel(level)
		if err != nil {
			return nil, err
		}

		opts = append(opts, log.FilterOption(filterFunc))
	default:
		opts = append(opts, log.LevelOption(logLvl))
	}

	return log.NewLogger(out, opts...), nil
}
