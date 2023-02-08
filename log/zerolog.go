package log

import (
	"os"
	"time"

	cmtlog "github.com/cometbft/cometbft/libs/log"
	"github.com/rs/zerolog"
)

// Defines commons keys for logging
const ModuleKey = "module"

var (
	// ContextKey is used to store the logger in the context
	ContextKey struct{}
	_          Logger = (*ZeroLogWrapper)(nil)
)

func NewZeroLogger(key, value string) *zerolog.Logger {
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.Kitchen}
	logger := zerolog.New(output).With().Str(key, value).Timestamp().Logger()
	return &logger
}

// TODO: add filtered logging in ZeroLog: https://github.com/cosmos/cosmos-sdk/pull/13236 / https://github.com/cosmos/cosmos-sdk/issues/13699#issuecomment-1354887644

// ZeroLogWrapper provides a wrapper around a zerolog.Logger instance. It implements
// Tendermint's Logger interface.
type ZeroLogWrapper struct {
	zerolog.Logger
}

// Info implements Tendermint's Logger interface and logs with level INFO. A set
// of key/value tuples may be provided to add context to the log. The number of
// tuples must be even and the key of the tuple must be a string.
func (z ZeroLogWrapper) Info(msg string, keyVals ...interface{}) {
	z.Logger.Info().Fields(getLogFields(keyVals...)).Msg(msg)
}

// Error implements Tendermint's Logger interface and logs with level ERR. A set
// of key/value tuples may be provided to add context to the log. The number of
// tuples must be even and the key of the tuple must be a string.
func (z ZeroLogWrapper) Error(msg string, keyVals ...interface{}) {
	z.Logger.Error().Fields(getLogFields(keyVals...)).Msg(msg)
}

// Debug implements Tendermint's Logger interface and logs with level DEBUG. A set
// of key/value tuples may be provided to add context to the log. The number of
// tuples must be even and the key of the tuple must be a string.
func (z ZeroLogWrapper) Debug(msg string, keyVals ...interface{}) {
	z.Logger.Debug().Fields(getLogFields(keyVals...)).Msg(msg)
}

// With returns a new wrapped logger with additional context provided by a set
// of key/value tuples. The number of tuples must be even and the key of the
// tuple must be a string.
func (z ZeroLogWrapper) With(keyVals ...interface{}) cmtlog.Logger {
	return ZeroLogWrapper{z.Logger.With().Fields(getLogFields(keyVals...)).Logger()}
}

func getLogFields(keyVals ...interface{}) map[string]interface{} {
	if len(keyVals)%2 != 0 {
		return nil
	}

	fields := make(map[string]interface{})
	for i := 0; i < len(keyVals); i += 2 {
		fields[keyVals[i].(string)] = keyVals[i+1]
	}

	return fields
}
