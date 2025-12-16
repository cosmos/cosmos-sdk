package log_test

import (
	"log/slog"
	"testing"

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
