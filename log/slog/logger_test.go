package slog_test

import (
	"bytes"
	"encoding/json"
	stdslog "log/slog"
	"testing"

	"cosmossdk.io/log/slog"
)

func TestSlog(t *testing.T) {
	var buf bytes.Buffer
	h := stdslog.NewJSONHandler(&buf, &stdslog.HandlerOptions{
		Level: stdslog.LevelDebug,
	})
	logger := slog.NewCustomLogger(stdslog.New(h))

	type logLine struct {
		Level string `json:"level"`
		Msg   string `json:"msg"`
		Num   int    `json:"num"`
	}

	var line logLine

	logger.Debug("Message one", "num", 1)
	if err := json.Unmarshal(buf.Bytes(), &line); err != nil {
		t.Fatal(err)
	}
	if want := (logLine{
		Level: stdslog.LevelDebug.String(),
		Msg:   "Message one",
		Num:   1,
	}); want != line {
		t.Fatalf("unexpected log record: want %v, got %v", want, line)
	}

	buf.Reset()
	logger.Info("Message two", "num", 2)
	if err := json.Unmarshal(buf.Bytes(), &line); err != nil {
		t.Fatal(err)
	}
	if want := (logLine{
		Level: stdslog.LevelInfo.String(),
		Msg:   "Message two",
		Num:   2,
	}); want != line {
		t.Fatalf("unexpected log record: want %v, got %v", want, line)
	}

	buf.Reset()
	logger.Warn("Message three", "num", 3)
	if err := json.Unmarshal(buf.Bytes(), &line); err != nil {
		t.Fatal(err)
	}
	if want := (logLine{
		Level: stdslog.LevelWarn.String(),
		Msg:   "Message three",
		Num:   3,
	}); want != line {
		t.Fatalf("unexpected log record: want %v, got %v", want, line)
	}

	buf.Reset()
	logger.Error("Message four", "num", 4)
	if err := json.Unmarshal(buf.Bytes(), &line); err != nil {
		t.Fatal(err)
	}
	if want := (logLine{
		Level: stdslog.LevelError.String(),
		Msg:   "Message four",
		Num:   4,
	}); want != line {
		t.Fatalf("unexpected log record: want %v, got %v", want, line)
	}

	wLogger := logger.With("num", 5)
	buf.Reset()
	wLogger.Info("Using .With")

	if err := json.Unmarshal(buf.Bytes(), &line); err != nil {
		t.Fatal(err)
	}
	if want := (logLine{
		Level: stdslog.LevelInfo.String(),
		Msg:   "Using .With",
		Num:   5,
	}); want != line {
		t.Fatalf("unexpected log record: want %v, got %v", want, line)
	}
}
