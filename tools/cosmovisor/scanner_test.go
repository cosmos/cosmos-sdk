package cosmovisor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	upgradetypes "cosmossdk.io/x/upgrade/types"
)

func TestParseUpgradeInfoFile(t *testing.T) {
	cases := []struct {
		filename      string
		expectUpgrade *upgradetypes.Plan
		disableRecase bool
		expectErr     string
	}{
		{
			filename:      "f1-good.json",
			disableRecase: false,
			expectUpgrade: &upgradetypes.Plan{Name: "upgrade1", Info: "some info", Height: 123},
		},
		{
			filename:      "f2-normalized-name.json",
			disableRecase: false,
			expectUpgrade: &upgradetypes.Plan{Name: "upgrade2", Info: "some info", Height: 125},
		},
		{
			filename:      "f2-normalized-name.json",
			disableRecase: true,
			expectUpgrade: &upgradetypes.Plan{Name: "Upgrade2", Info: "some info", Height: 125},
		},
		{
			filename:      "f2-bad-type.json",
			disableRecase: false,
			expectUpgrade: nil,
			expectErr:     "cannot unmarshal number into Go value",
		},
		{
			filename:      "f2-bad-type-2.json",
			disableRecase: false,
			expectUpgrade: nil,
			expectErr:     `unknown field "heigh"`,
		},
		{
			filename:      "f3-empty.json",
			disableRecase: false,
			expectUpgrade: nil,
			expectErr:     "EOF",
		},
		{
			filename:      "f4-empty-obj.json",
			disableRecase: false,
			expectUpgrade: nil,
			expectErr:     "name cannot be empty",
		},
		{
			filename:      "f5-partial-obj-1.json",
			disableRecase: false,
			expectUpgrade: nil,
			expectErr:     "height must be greater than 0",
		},
		{
			filename:      "f5-partial-obj-2.json",
			disableRecase: false,
			expectUpgrade: nil,
			expectErr:     "name cannot be empty: invalid request",
		},
		{
			filename:      "non-existent.json",
			disableRecase: false,
			expectUpgrade: nil,
			expectErr:     "no such file or directory",
		},
	}

	for i := range cases {
		tc := cases[i]
		t.Run(tc.filename, func(t *testing.T) {
			require := require.New(t)
			ui, err := parseUpgradeInfoFile(filepath.Join(".", "testdata", "upgrade-files", tc.filename), tc.disableRecase)
			if tc.expectErr != "" {
				require.Error(err)
				require.Contains(err.Error(), tc.expectErr)
			} else {
				require.NoError(err)
				require.Equal(tc.expectUpgrade, ui)
			}
		})
	}
}

func parseUpgradeInfoFile(filename string, disableRecase bool) (*upgradetypes.Plan, error) {
	cfg := &Config{
		DisableRecase: disableRecase,
	}
	bz, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return cfg.ParseUpgradeInfo(bz)
}
