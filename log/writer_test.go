package log_test

import (
	"bytes"
	"strings"
	"testing"

	"cosmossdk.io/log"
)

func TestFilteredWriter(t *testing.T) {
	buf := new(bytes.Buffer)

	level := "consensus:debug,mempool:debug,*:error"
	filter, err := log.ParseLogLevel(level)
	if err != nil {
		t.Fatalf("failed to parse log level: %v", err)
	}

	logger := log.NewLogger(buf, log.FilterOption(filter))
	logger.Debug("this log line should be displayed", log.ModuleKey, "consensus")
	if !strings.Contains(buf.String(), "this log line should be displayed") {
		t.Errorf("expected log line to be displayed, but it was not")
	}
	buf.Reset()

	logger.Debug("this log line should be filtered", log.ModuleKey, "server")
	if buf.Len() != 0 {
		t.Errorf("expected log line to be filtered, but it was not")
	}
}
