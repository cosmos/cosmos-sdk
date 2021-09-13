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

func TestGetVersion(t *testing.T) {
	cases := []struct {
		name        string
		versionStr  string
		expectValid bool
		expectRes   string
	}{
		{
			name:        "valid version string",
			versionStr:  "v1.0.0",
			expectValid: true,
			expectRes:   "v1.0.0",
		}, {
			name:        "valid git tag string",
			versionStr:  "v1.1.0-alpha2-1-g81f1347e",
			expectValid: true,
			expectRes:   "v1.1.0",
		}, {
			name:        "invalid string",
			versionStr:  "v1.test",
			expectValid: false,
			expectRes:   "",
		}, {
			name:        "incomplete version string",
			versionStr:  "v1.1",
			expectValid: false,
			expectRes:   "",
		}, {
			name:        "incomplete git tag string",
			versionStr:  "v1.1-alpha2-1-g81f1347e",
			expectValid: false,
			expectRes:   "",
		},
	}

	for i := range cases {
		tc := cases[i]
		t.Run(tc.name, func(t *testing.T) {
			require := require.New(t)
			isValid, res := getVersion(tc.versionStr)
			if tc.expectValid {
				require.True(isValid)
				require.Equal(tc.expectRes, res)
			} else {
				require.False(isValid)
			}
		})
	}
}
