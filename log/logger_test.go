package log_test

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"

	"cosmossdk.io/log/v2"
)

func TestLoggerBasic(t *testing.T) {
	buf := new(bytes.Buffer)
	logger := log.NewLogger("test", log.WithConsoleWriter(buf), log.WithColor(false))
	logger.Info("hello world")
	if !strings.Contains(buf.String(), "hello world") {
		t.Fatalf("expected hello world, got: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "INF") {
		t.Fatalf("expected INF level, got: %s", buf.String())
	}
}

func TestLoggerWithKeyVals(t *testing.T) {
	buf := new(bytes.Buffer)
	logger := log.NewLogger("test", log.WithConsoleWriter(buf), log.WithColor(false))
	logger.Info("hello world", "key1", "value1", "key2", 42)
	output := buf.String()
	if !strings.Contains(output, "hello world") {
		t.Fatalf("expected hello world, got: %s", output)
	}
	if !strings.Contains(output, "key1=value1") {
		t.Fatalf("expected key1=value1, got: %s", output)
	}
	if !strings.Contains(output, "key2=42") {
		t.Fatalf("expected key2=42, got: %s", output)
	}
}

func TestLoggerWith(t *testing.T) {
	buf := new(bytes.Buffer)
	logger := log.NewLogger("test", log.WithConsoleWriter(buf), log.WithColor(false))
	logger = logger.With("module", "test-module")
	logger.Info("hello world")
	output := buf.String()
	if !strings.Contains(output, "module=test-module") {
		t.Fatalf("expected module=test-module, got: %s", output)
	}
}

func TestLoggerLevels(t *testing.T) {
	tests := []struct {
		name     string
		logFunc  func(log.Logger)
		expected string
	}{
		{"debug", func(l log.Logger) { l.Debug("test") }, "DBG"},
		{"info", func(l log.Logger) { l.Info("test") }, "INF"},
		{"warn", func(l log.Logger) { l.Warn("test") }, "WRN"},
		{"error", func(l log.Logger) { l.Error("test") }, "ERR"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			logger := log.NewLogger("test",
				log.WithConsoleWriter(buf),
				log.WithColor(false),
				log.WithLevel(slog.LevelDebug),
			)
			tc.logFunc(logger)
			if !strings.Contains(buf.String(), tc.expected) {
				t.Fatalf("expected %s level, got: %s", tc.expected, buf.String())
			}
		})
	}
}

func TestLoggerJSON(t *testing.T) {
	buf := new(bytes.Buffer)
	logger := log.NewLogger("test", log.WithConsoleWriter(buf), log.WithJSONOutput())
	logger.Info("hello world", "key", "value")
	output := buf.String()
	// zerolog JSON uses "message" field
	if !strings.Contains(output, `"message":"hello world"`) {
		t.Fatalf("expected JSON output with message field, got: %s", output)
	}
	if !strings.Contains(output, `"key":"value"`) {
		t.Fatalf("expected JSON output with key field, got: %s", output)
	}
}

func TestFilteredWriter(t *testing.T) {
	buf := new(bytes.Buffer)

	level := "consensus:debug,mempool:debug,*:error"
	filter, err := log.ParseLogLevel(level)
	if err != nil {
		t.Fatalf("failed to parse log level: %v", err)
	}

	logger := log.NewLogger("test",
		log.WithConsoleWriter(buf),
		log.WithColor(false),
		log.WithLevel(slog.LevelDebug),
		log.WithFilter(filter),
	)

	logger.Debug("this log line should be displayed", log.ModuleKey, "consensus")
	if !strings.Contains(buf.String(), "this log line should be displayed") {
		t.Errorf("expected log line to be displayed, but it was not: %s", buf.String())
	}
	buf.Reset()

	logger.Debug("this log line should be filtered", log.ModuleKey, "server")
	if buf.Len() != 0 {
		t.Errorf("expected log line to be filtered, but it was not: %s", buf.String())
	}
}

func TestNopLogger(t *testing.T) {
	logger := log.NewNopLogger()
	// Should not panic
	logger.Info("test")
	logger.Debug("test")
	logger.Warn("test")
	logger.Error("test")
	logger.With("key", "value").Info("test")
}

func TestWithoutConsole(t *testing.T) {
	buf := new(bytes.Buffer)
	// WithoutConsole should suppress all console output (OTEL disabled path)
	logger := log.NewLogger("test",
		log.WithConsoleWriter(buf),
		log.WithoutConsole(),
	)
	logger.Info("this should not appear")
	logger.Warn("neither should this")
	if buf.Len() != 0 {
		t.Errorf("expected no output with WithoutConsole, got: %s", buf.String())
	}
}
