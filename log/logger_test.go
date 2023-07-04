package log_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"cosmossdk.io/log"
)

func TestLoggerOptionStackTrace(t *testing.T) {
	t.Skip() // todo(@julienrbrt) unskip when https://github.com/rs/zerolog/pull/560 merged

	buf := new(bytes.Buffer)
	logger := log.NewLogger(buf, log.TraceOption(true), log.ColorOption(false))
	logger.Error("this log should be displayed", "error", inner())
	if strings.Count(buf.String(), "logger_test.go") != 1 {
		t.Fatalf("stack trace not found, got: %s", buf.String())
	}
	buf.Reset()

	logger = log.NewLogger(buf, log.TraceOption(false), log.ColorOption(false))
	logger.Error("this log should be displayed", "error", inner())
	if strings.Count(buf.String(), "logger_test.go") > 0 {
		t.Fatalf("stack trace found, got: %s", buf.String())
	}
}

func inner() error {
	return errors.New("seems we have an error here")
}
