package logmonitor

import (
	"bytes"
	"strings"
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

			lm := NewLogMonitor(shutdownFn, tc.shutdownStrings)

			n, err := lm.Write([]byte(tc.input))
			require.NoError(t, err)
			require.Equal(t, len(tc.input), n)

			require.Equal(t, tc.expectShutdown, shutdownCalled)
		})
	}
}

func TestMultiWriter(t *testing.T) {
	shutdownCalled := false
	shutdownFn := func(reason string) {
		shutdownCalled = true
	}
	shutdownStrings := []string{"CONSENSUS FAILURE!"}

	lm := NewLogMonitor(shutdownFn, shutdownStrings)

	buf1 := &bytes.Buffer{}
	buf2 := &bytes.Buffer{}

	mw := NewMultiWriter(lm, buf1, buf2)

	testString := "Test log message"
	n, err := mw.Write([]byte(testString))

	require.NoError(t, err)
	require.Equal(t, len(testString), n)
	require.Equal(t, testString, buf1.String())
	require.Equal(t, testString, buf2.String())
	require.False(t, shutdownCalled)

	criticalError := "CONSENSUS FAILURE! Critical error"
	_, _ = mw.Write([]byte(criticalError))

	require.True(t, shutdownCalled)
	require.True(t, strings.Contains(buf1.String(), criticalError))
	require.True(t, strings.Contains(buf2.String(), criticalError))
}
