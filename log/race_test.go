package log_test

import (
	"io"
	"testing"

	log "cosmossdk.io/log/v2"
)

func TestSetVerboseModeRace(t *testing.T) {
	logger := log.NewLogger(io.Discard, log.FilterOption(func(module, level string) bool {
		return false
	}))
	vl := logger.(log.VerboseModeLogger)

	done := make(chan struct{})

	go func() {
		for i := 0; i < 100000; i++ {
			logger.Info("msg")
		}
		close(done)
	}()

	for i := 0; i < 100000; i++ {
		vl.SetVerboseMode(true)
		vl.SetVerboseMode(false)
	}

	<-done
}
