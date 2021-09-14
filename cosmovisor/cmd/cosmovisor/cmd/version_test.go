package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsVersionCommand(t *testing.T) {
	cases := []struct {
		name      string
		args      []string
		expectRes bool
	}{{
		name:      "valid args - lowercase",
		args:      []string{"version"},
		expectRes: true,
	}, {
		name:      "typo",
		args:      []string{"vrsion"},
		expectRes: false,
	}, {
		name:      "non version command",
		args:      []string{"start"},
		expectRes: false,
	}, {
		name:      "incorrect format",
		args:      []string{"start", "version"},
		expectRes: false,
	}}

	for i := range cases {
		tc := cases[i]
		t.Run(tc.name, func(t *testing.T) {
			require := require.New(t)
			res := isVersionCommand(tc.args)
			require.Equal(tc.expectRes, res)
		})
	}
}
