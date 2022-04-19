package cmd

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsRunCommand(t *testing.T) {
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
			name:     "run",
			arg:      "run",
			expected: true,
		},
		{
			name:     "run weird casing",
			arg:      "RUn",
			expected: true,
		},
		{
			name:     "--run",
			arg:      "--run",
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
			name:     "version",
			arg:      "version",
			expected: false,
		},
		{
			name:     "--version",
			arg:      "--version",
			expected: false,
		},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("%s - %t", tc.name, tc.expected), func(t *testing.T) {
			actual := IsRunCommand(tc.arg)
			require.Equal(t, tc.expected, actual)
		})
	}
}

// TODO: Write tests for func Run(args []string) error
