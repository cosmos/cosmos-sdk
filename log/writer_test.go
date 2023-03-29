package log_test

import (
	"bytes"
	"strings"
	"testing"

	"cosmossdk.io/log"
	"gotest.tools/v3/assert"
)

func TestFilteredWriter(t *testing.T) {
	checkBuf := new(bytes.Buffer)

	level := "consensus:debug,mempool:debug,*:error"
	filter, err := log.ParseLogLevel(level)
	assert.NilError(t, err)

	logger := log.NewLogger(checkBuf, log.FilterOption(filter))
	logger.Debug("this log line should be displayed", log.ModuleKey, "consensus")
	assert.Check(t, strings.Contains(checkBuf.String(), "this log line should be displayed"))
	checkBuf.Reset()

	logger.Debug("this log line should be filtered", log.ModuleKey, "server")
	assert.Check(t, !strings.Contains(checkBuf.String(), "this log line should be filtered"))
	checkBuf.Reset()
}
