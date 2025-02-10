package cosmovisor

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	upgradetypes "cosmossdk.io/x/upgrade/types"
)

func TestParseUpgradeInfoFile(t *testing.T) {
	cases := []struct {
		filename      string
		expectUpgrade upgradetypes.Plan
		expectErr     bool
	}{
		{
			filename:      "f1-good.json",
			expectUpgrade: upgradetypes.Plan{Name: "upgrade1", Info: "some info", Height: 123},
			expectErr:     false,
		},
		{
			filename:      "f2-normalized-name.json",
			expectUpgrade: upgradetypes.Plan{Name: "upgrade2", Info: "some info", Height: 125},
			expectErr:     false,
		},
		{
			filename:      "f2-bad-type.json",
			expectUpgrade: upgradetypes.Plan{},
			expectErr:     true,
		},
		{
			filename:      "f2-bad-type-2.json",
			expectUpgrade: upgradetypes.Plan{},
			expectErr:     true,
		},
		{
			filename:      "f3-empty.json",
			expectUpgrade: upgradetypes.Plan{},
			expectErr:     true,
		},
		{
			filename:      "f4-empty-obj.json",
			expectUpgrade: upgradetypes.Plan{},
			expectErr:     true,
		},
		{
			filename:      "f5-partial-obj-1.json",
			expectUpgrade: upgradetypes.Plan{},
			expectErr:     true,
		},
		{
			filename:      "f5-partial-obj-2.json",
			expectUpgrade: upgradetypes.Plan{},
			expectErr:     true,
		},
		{
			filename:      "unknown.json",
			expectUpgrade: upgradetypes.Plan{},
			expectErr:     true,
		},
	}

	for i := range cases {
		tc := cases[i]
		t.Run(tc.filename, func(t *testing.T) {
			require := require.New(t)
			ui, err := parseUpgradeInfoFile(filepath.Join(".", "testdata", "upgrade-files", tc.filename))
			if tc.expectErr {
				require.Error(err)
			} else {
				require.NoError(err)
				require.Equal(tc.expectUpgrade, ui)
			}
		})
	}
}
