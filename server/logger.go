package server

import (
	"github.com/rs/zerolog"
	tmlog "github.com/tendermint/tendermint/libs/log"
)

var _ tmlog.Logger = (*ZeroLogWrapper)(nil)

type ZeroLogWrapper struct {
	zerolog.Logger
}

func (z ZeroLogWrapper) Info(msg string, keyVals ...interface{}) {
	z.Logger.Info().Fields(getLogFields(keyVals)).Msg(msg)
}

func (z ZeroLogWrapper) Error(msg string, keyVals ...interface{}) {
	z.Logger.Error().Fields(getLogFields(keyVals)).Msg(msg)
}

func (z ZeroLogWrapper) Debug(msg string, keyVals ...interface{}) {
	z.Logger.Debug().Fields(getLogFields(keyVals)).Msg(msg)
}

func (z ZeroLogWrapper) With(keyVals ...interface{}) tmlog.Logger {
	return ZeroLogWrapper{z.Logger.With().Fields(getLogFields(keyVals)).Logger()}
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
