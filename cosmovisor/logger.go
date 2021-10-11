package cosmovisor

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

var Logger zerolog.Logger

func SetupLogging() {
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.Kitchen}
	Logger = zerolog.New(output).With().Str("module", "cosmovisor").Timestamp().Logger()
}
