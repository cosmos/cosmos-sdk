package server

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  slog.Level
		expectErr bool
	}{
		// Valid levels
		{"debug lowercase", "debug", slog.LevelDebug, false},
		{"debug uppercase", "DEBUG", slog.LevelDebug, false},
		{"debug mixed case", "Debug", slog.LevelDebug, false},
		{"info lowercase", "info", slog.LevelInfo, false},
		{"info uppercase", "INFO", slog.LevelInfo, false},
		{"empty string defaults to info", "", slog.LevelInfo, false},
		{"warn lowercase", "warn", slog.LevelWarn, false},
		{"warning alias", "warning", slog.LevelWarn, false},
		{"warn uppercase", "WARN", slog.LevelWarn, false},
		{"error lowercase", "error", slog.LevelError, false},
		{"err alias", "err", slog.LevelError, false},
		{"error uppercase", "ERROR", slog.LevelError, false},

		// Invalid levels
		{"invalid level", "invalid", 0, true},
		{"trace not supported", "trace", 0, true},
		{"fatal not supported", "fatal", 0, true},
		{"panic not supported", "panic", 0, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			level, err := parseLogLevel(tc.input)

			if tc.expectErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), "unknown log level")
				require.Contains(t, err.Error(), "valid levels")
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, level)
			}
		})
	}
}
