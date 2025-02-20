package cosmovisor

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

var LoggerKey struct{}

func NewLogger() *zerolog.Logger {
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.Kitchen}
	logger := zerolog.New(output).With().Str("module", "cosmovisor").Timestamp().Logger()
	return &logger
}
