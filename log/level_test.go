package log_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/rs/zerolog"

	"cosmossdk.io/log"
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
	if filter("server", "panic") {
		t.Errorf("expected filter to return false for server:panic")
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
	if filter == nil {
		t.Fatalf("expected non-nil filter, got nil")
	}

	if filter("state", "error") {
		t.Errorf("expected filter to return false for state:error")
	}
	if filter("consensus", "error") {
		t.Errorf("expected filter to return false for consensus:error")
	}

	if !filter("consensus", "debug") {
		t.Errorf("expected filter to return true for consensus:debug")
	}
	if !filter("consensus", "info") {
		t.Errorf("expected filter to return true for consensus:info")
	}
	if !filter("state", "debug") {
		t.Errorf("expected filter to return true for state:debug")
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
		level        zerolog.Level
		verboseLevel zerolog.Level
		filter       string
		expected     string
	}{
		{
			name:         "verbose mode simple case",
			level:        zerolog.WarnLevel,
			verboseLevel: zerolog.DebugLevel,
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
			level:        zerolog.WarnLevel,
			verboseLevel: zerolog.InfoLevel,
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
			level:        zerolog.WarnLevel,
			verboseLevel: zerolog.NoLevel,
			expected: `* WRN msg 2 module=foo
* ERR msg 3 module=bar
* ERR Start Verbose Mode
* WRN msg 2 module=foo
* ERR msg 3 module=bar
`,
		},
		{
			name:         "no verbose mode with filter",
			level:        zerolog.WarnLevel,
			verboseLevel: zerolog.NoLevel,
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
				log.LevelOption(tc.level),
				log.VerboseLevelOption(tc.verboseLevel),
				log.ColorOption(false),
				log.TimeFormatOption("*"), // disable non-deterministic time format
			}
			if tc.filter != "" {
				filter, err := log.ParseLogLevel(tc.filter)
				if err != nil {
					t.Fatalf("failed to parse log level: %v", err)
				}
				opts = append(opts, log.FilterOption(filter))
			}
			logger := log.NewLogger(out, opts...)
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
