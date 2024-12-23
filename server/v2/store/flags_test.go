package store

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrefix(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"test", "store.test"},
		{"", "store."},
		{"app-db-backend", "store.app-db-backend"},
	}

	for _, tt := range tests {
		result := prefix(tt.input)
		require.Equal(t, tt.expected, result)
	}
}

func TestFlagConstants(t *testing.T) {
	require.Equal(t, "store.app-db-backend", FlagAppDBBackend)
	require.Equal(t, "store.keep-recent", FlagKeepRecent)
	require.Equal(t, "store.interval", FlagInterval)
}
