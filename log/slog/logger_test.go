package slog_test

import (
	"bytes"
	"encoding/json"
	stdslog "log/slog"
	"testing"

	"cosmossdk.io/log/slog"
	"github.com/stretchr/testify/require"
)

func TestSlog(t *testing.T) {
	var buf bytes.Buffer
	h := stdslog.NewJSONHandler(&buf, &stdslog.HandlerOptions{
		Level: stdslog.LevelDebug,
	})
	logger := slog.FromSlog(stdslog.New(h))

	type logLine struct {
		Level string `json:"level"`
		Msg   string `json:"msg"`
		Num   int    `json:"num"`
	}

	var line logLine

	logger.Debug("Message one", "num", 1)
	require.NoError(t, json.Unmarshal(buf.Bytes(), &line))
	require.Equal(t, logLine{
		Level: stdslog.LevelDebug.String(),
		Msg:   "Message one",
		Num:   1,
	}, line)

	buf.Reset()
	logger.Info("Message two", "num", 2)
	require.NoError(t, json.Unmarshal(buf.Bytes(), &line))
	require.Equal(t, logLine{
		Level: stdslog.LevelInfo.String(),
		Msg:   "Message two",
		Num:   2,
	}, line)

	buf.Reset()
	logger.Warn("Message three", "num", 3)
	require.NoError(t, json.Unmarshal(buf.Bytes(), &line))
	require.Equal(t, logLine{
		Level: stdslog.LevelWarn.String(),
		Msg:   "Message three",
		Num:   3,
	}, line)

	buf.Reset()
	logger.Error("Message four", "num", 4)
	require.NoError(t, json.Unmarshal(buf.Bytes(), &line))
	require.Equal(t, logLine{
		Level: stdslog.LevelError.String(),
		Msg:   "Message four",
		Num:   4,
	}, line)
}
