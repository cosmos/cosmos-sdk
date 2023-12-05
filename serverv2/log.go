package serverv2

import (
	"io"

	"github.com/rs/zerolog"
	"github.com/spf13/viper"

	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/client/flags"
)

// NewLogger creates a the default SDK logger.
// It reads the log level and format from the server context.
func NewLogger(v *viper.Viper, out io.Writer) (log.Logger, error) {
	var opts []log.Option
	if v.GetString(flags.FlagLogFormat) == flags.OutputFormatJSON {
		opts = append(opts, log.OutputJSONOption())
	}

	// check and set filter level or keys for the logger if any
	logLvlStr := v.GetString(flags.FlagLogLevel)
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

	// Check if the flag for trace logging is set and enable stack traces if so.
	opts = append(opts, log.TraceOption(viper.GetBool("trace")))

	return log.NewLogger(out, opts...).With(log.ModuleKey, "server"), nil
}
