package serverv2

import (
	"io"

	"github.com/rs/zerolog"
	"github.com/spf13/viper"

	"cosmossdk.io/log"
)

// NewLogger creates the default SDK logger.
// It reads the log level and format from the server context.
func NewLogger(v *viper.Viper, out io.Writer) (log.Logger, error) {
	var opts []log.Option
	if v.GetString(FlagLogFormat) == OutputFormatJSON {
		opts = append(opts, log.OutputJSONOption())
	}
	opts = append(opts,
		log.ColorOption(!v.GetBool(FlagLogNoColor)),
		log.TraceOption(v.GetBool(FlagTrace)))

	// check and set filter level or keys for the logger if any
	logLvlStr := v.GetString(FlagLogLevel)
	if logLvlStr == "" {
		return log.NewLogger(out, opts...), nil
	}

	logLvl, err := zerolog.ParseLevel(logLvlStr)
	switch {
	case err != nil:
		// If the log level is not a valid zerolog level, then we try to parse it as a key filter.
		filterFunc, err := log.ParseLogLevel(logLvlStr)
		if err != nil {
			return nil, err
		}

		opts = append(opts, log.FilterOption(filterFunc))
	default:
		opts = append(opts, log.LevelOption(logLvl))
	}

	return log.NewLogger(out, opts...), nil
}
