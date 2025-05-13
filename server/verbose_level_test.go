package server

import (
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestParseVerboseLogLevel(t *testing.T) {
	tt := []struct {
		input    string
		expected zerolog.Level
	}{
		// mainly testing that none maps to NoLevel, but checking other cases too for sanity
		{"none", zerolog.NoLevel},
		{"debug", zerolog.DebugLevel},
		{"info", zerolog.InfoLevel},
		{"warn", zerolog.WarnLevel},
		{"error", zerolog.ErrorLevel},
		{"fatal", zerolog.FatalLevel},
		{"panic", zerolog.PanicLevel},
		{"trace", zerolog.TraceLevel},
		{"disabled", zerolog.Disabled},
	}

	for _, tc := range tt {
		lvl, err := parseVerboseLogLevel(tc.input)
		require.NoError(t, err)
		require.Equal(t, tc.expected, lvl)
	}
}
