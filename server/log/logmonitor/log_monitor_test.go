package logmonitor

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLogMonitorWrite(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		shutdownStrings []string
		expectShutdown  bool
	}{
		{
			name:            "normal log",
			input:           "This is a normal log message",
			shutdownStrings: []string{"CONSENSUS FAILURE!"},
			expectShutdown:  false,
		},
		{
			name:            "shutdown trigger",
			input:           "Error: CONSENSUS FAILURE! Unable to proceed.",
			shutdownStrings: []string{"CONSENSUS FAILURE!"},
			expectShutdown:  true,
		},
		{
			name:            "multiple shutdown strings",
			input:           "Warning: CRITICAL ERROR detected",
			shutdownStrings: []string{"CONSENSUS FAILURE!", "CRITICAL ERROR"},
			expectShutdown:  true,
		},
		{
			name:            "case insensitive",
			input:           "Error: Consensus Failure! detected",
			shutdownStrings: []string{"CONSENSUS FAILURE!"},
			expectShutdown:  false,
		},
		{
			name:            "with ANSI color codes",
			input:           "\x1b[31mError: CONSENSUS FAILURE! Unable to proceed.\x1b[0m",
			shutdownStrings: []string{"CONSENSUS FAILURE!"},
			expectShutdown:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			shutdownCalled := false
			shutdownFn := func(reason string) {
				shutdownCalled = true
			}

			cfg := &Config{ShutdownStrings: tc.shutdownStrings}
			lm := NewLogMonitor(cfg, shutdownFn)

			n, err := lm.Write([]byte(tc.input))
			require.NoError(t, err)
			require.Equal(t, len(tc.input), n)

			require.Equal(t, tc.expectShutdown, shutdownCalled)
		})
	}
}

func TestInitGlobalLogMonitor(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Enabled = true

	shutdownCalled := false
	shutdownFn := func(reason string) {
		shutdownCalled = true
	}

	stdout, stderr := InitGlobalLogMonitor(cfg, shutdownFn)

	require.NotNil(t, stdout)
	require.NotNil(t, stderr)

	_, err := stdout.Write([]byte("CONSENSUS FAILURE!\n"))
	require.NoError(t, err)

	require.True(t, shutdownCalled)
}
