package log

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

// Defines commons keys for logging
const ModuleKey = "module"

var LoggerKey struct{}

func NewLogger(key, value string) *zerolog.Logger {
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.Kitchen}
	logger := zerolog.New(output).With().Str(key, value).Timestamp().Logger()
	return &logger
}
