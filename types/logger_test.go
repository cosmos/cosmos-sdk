package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMemLogger(t *testing.T) {
	logger := NewMemLogger()
	logger.Info("msg")
	require.Equal(t, logger.Logs(), []LogEntry{LogEntry{"info", "msg", nil}})
	logger.Debug("msg2")
	require.Equal(t, logger.Logs(), []LogEntry{LogEntry{"info", "msg", nil}, LogEntry{"debug", "msg2", nil}})
	logger.Error("msg3", 2)
	require.Equal(t, logger.Logs(), []LogEntry{LogEntry{"info", "msg", nil}, LogEntry{"debug", "msg2", nil}, LogEntry{"error", "msg3", []interface{}{2}}})
}
