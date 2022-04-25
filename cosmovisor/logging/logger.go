package logging

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

type Logger struct {
	zerolog.Logger
}

func NewLogger() *Logger {
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.Kitchen}
	return &Logger{zerolog.New(output).With().Str("module", "cosmovisor").Timestamp().Logger()}
}

func (l *Logger) DisableLogger() {
	l.Logger = l.Logger.Level(zerolog.Disabled)
}

func (l *Logger) EnableLogger() {
	l.Logger = l.Logger.Level(zerolog.DebugLevel)
}
