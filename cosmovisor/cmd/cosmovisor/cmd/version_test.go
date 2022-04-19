package cmd

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsVersionCommand(t *testing.T) {
	cases := []struct {
		name     string
		arg      string
		expected bool
	}{
		{
			name:     "empty string",
			arg:      "",
			expected: false,
		},
		{
			name:     "random",
			arg:      "random",
			expected: false,
		},
		{
			name:     "version",
			arg:      "version",
			expected: true,
		},
		{
			name:     "--version",
			arg:      "--version",
			expected: true,
		},
		{
			name:     "version weird casing",
			arg:      "veRSiOn",
			expected: true,
		},
		{
			// -v should be reserved for verbose, and should not be used for --version.
			name:     "-v",
			arg:      "-v",
			expected: false,
		},
		{
			name:     "typo",
			arg:      "vrsion",
			expected: false,
		},
		{
			name:     "non version command",
			arg:      "start",
			expected: false,
		},
		{
			name:     "help",
			arg:      "help",
			expected: false,
		},
		{
			name:     "-h",
			arg:      "-h",
			expected: false,
		},
		{
			name:     "--help",
			arg:      "--help",
			expected: false,
		},
		{
			name:     "run",
			arg:      "run",
			expected: false,
		},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("%s - %t", tc.name, tc.expected), func(t *testing.T) {
			actual := IsVersionCommand(tc.arg)
			require.Equal(t, tc.expected, actual)
		})
	}
}
