package log_test

import (
	"bytes"
	"fmt"
	"log/slog"
	"testing"

	"cosmossdk.io/log"
	"github.com/rs/zerolog"
)

func TestParseLogLevel(t *testing.T) {
	_, err := log.ParseLogLevel("")
	if err == nil {
		t.Errorf("expected error for empty log level, got nil")
	}

	level := "consensus:foo,mempool:debug,*:error"
	_, err = log.ParseLogLevel(level)
	if err == nil {
		t.Errorf("expected error for invalid log level foo in log level list [consensus:foo mempool:debug *:error], got nil")
	}

	level = "consensus:debug,mempool:debug,*:error"
	filter, err := log.ParseLogLevel(level)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if filter == nil {
		t.Fatalf("expected non-nil filter, got nil")
	}

	if filter("consensus", "debug") {
		t.Errorf("expected filter to return false for consensus:debug")
	}
	if filter("consensus", "info") {
		t.Errorf("expected filter to return false for consensus:info")
	}
	if filter("consensus", "error") {
		t.Errorf("expected filter to return false for consensus:error")
	}
	if filter("mempool", "debug") {
		t.Errorf("expected filter to return false for mempool:debug")
	}
	if filter("mempool", "info") {
		t.Errorf("expected filter to return false for mempool:info")
	}
	if filter("mempool", "error") {
		t.Errorf("expected filter to return false for mempool:error")
	}
	if filter("state", "error") {
		t.Errorf("expected filter to return false for state:error")
	}

	if !filter("server", "debug") {
		t.Errorf("expected filter to return true for server:debug")
	}
	if !filter("state", "debug") {
		t.Errorf("expected filter to return true for state:debug")
	}
	if !filter("state", "info") {
		t.Errorf("expected filter to return true for state:info")
	}

	level = "error"
	filter, err = log.ParseLogLevel(level)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// For simple levels, filter should be nil
	if filter != nil {
		t.Fatalf("expected nil filter for simple level, got non-nil")
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected slog.Level
		wantErr  bool
	}{
		{"debug", slog.LevelDebug, false},
		{"DEBUG", slog.LevelDebug, false},
		{"info", slog.LevelInfo, false},
		{"INFO", slog.LevelInfo, false},
		{"warn", slog.LevelWarn, false},
		{"warning", slog.LevelWarn, false},
		{"error", slog.LevelError, false},
		{"err", slog.LevelError, false},
		{"disabled", slog.Level(100), false},
		{"none", slog.Level(100), false},
		{"invalid", slog.LevelInfo, true},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			level, err := log.ParseLevel(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error for %s, got nil", tc.input)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error for %s: %v", tc.input, err)
				return
			}
			if level != tc.expected {
				t.Errorf("expected %v for %s, got %v", tc.expected, tc.input, level)
			}
		})
	}
}

func TestLevelToString(t *testing.T) {
	tests := []struct {
		level    slog.Level
		expected string
	}{
		{slog.LevelDebug, "debug"},
		{slog.LevelInfo, "info"},
		{slog.LevelWarn, "warn"},
		{slog.LevelError, "error"},
		{slog.Level(100), "disabled"},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			result := log.LevelToString(tc.level)
			if result != tc.expected {
				t.Errorf("expected %s for level %v, got %s", tc.expected, tc.level, result)
			}
		})
	}
}

func TestVerboseMode(t *testing.T) {
	logMessages := []struct {
		level   zerolog.Level
		module  string
		message string
	}{
		{
			zerolog.InfoLevel,
			"foo",
			"msg 1",
		},
		{
			zerolog.WarnLevel,
			"foo",
			"msg 2",
		},
		{
			zerolog.ErrorLevel,
			"bar",
			"msg 3",
		},
		{
			zerolog.DebugLevel,
			"foo",
			"msg 4",
		},
	}
	tt := []struct {
		name         string
		level        slog.Level
		verboseLevel slog.Level
		filter       string
		expected     string
	}{
		{
			name:         "verbose mode simple case",
			level:        slog.LevelWarn,
			verboseLevel: slog.LevelDebug,
			expected: `* WRN msg 2 module=foo
* ERR msg 3 module=bar
* ERR Start Verbose Mode
* INF msg 1 module=foo
* WRN msg 2 module=foo
* ERR msg 3 module=bar
* DBG msg 4 module=foo
`,
		},
		{
			name:         "verbose mode with filter",
			level:        slog.LevelWarn,
			verboseLevel: slog.LevelInfo,
			filter:       "foo:error",
			expected: `* ERR msg 3 module=bar
* ERR Start Verbose Mode
* INF msg 1 module=foo
* WRN msg 2 module=foo
* ERR msg 3 module=bar
`,
		},
		{
			name:         "no verbose mode",
			level:        slog.LevelWarn,
			verboseLevel: log.NoLevel, // meant to be no level
			expected: `* WRN msg 2 module=foo
* ERR msg 3 module=bar
* ERR Start Verbose Mode
* WRN msg 2 module=foo
* ERR msg 3 module=bar
`,
		},
		{
			name:         "no verbose mode with filter",
			level:        slog.LevelWarn,
			verboseLevel: log.NoLevel, // meant to be no level
			filter:       "foo:error",
			expected: `* ERR msg 3 module=bar
* ERR Start Verbose Mode
* ERR msg 3 module=bar
`,
		},
	}
	for i, tc := range tt {
		t.Run(fmt.Sprintf("%d: %s", i, tc.name), func(t *testing.T) {
			out := new(bytes.Buffer)
			opts := []log.Option{
				log.WithLevel(tc.level),
				log.WithVerboseLevel(tc.verboseLevel),
				log.WithColor(false),
				log.WithTimeFormat("*"), // disable non-deterministic time format
			}
			if tc.filter != "" {
				filter, err := log.ParseLogLevel(tc.filter)
				if err != nil {
					t.Fatalf("failed to parse log level: %v", err)
				}
				opts = append(opts, log.WithFilter(filter))
			}
			opts = append(opts, log.WithConsoleWriter(out))
			logger := log.NewLogger("test", opts...)
			writeMsgs := func() {
				for _, msg := range logMessages {
					switch msg.level {
					case zerolog.InfoLevel:
						logger.Info(msg.message, log.ModuleKey, msg.module)
					case zerolog.WarnLevel:
						logger.Warn(msg.message, log.ModuleKey, msg.module)
					case zerolog.DebugLevel:
						logger.Debug(msg.message, log.ModuleKey, msg.module)
					case zerolog.ErrorLevel:
						logger.Error(msg.message, log.ModuleKey, msg.module)
					default:
						t.Fatalf("unexpected level: %v", msg.level)
					}
				}
			}
			writeMsgs()
			logger.Error("Start Verbose Mode")
			logger.(log.VerboseModeLogger).SetVerboseMode(true)
			writeMsgs()
			if tc.expected != out.String() {
				t.Fatalf("expected:\n%s\ngot:\n%s", tc.expected, out.String())
			}
		})
	}
}
