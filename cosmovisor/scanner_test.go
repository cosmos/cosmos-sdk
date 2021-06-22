package cosmovisor

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseUpgradeInfoFile(t *testing.T) {
	cases := []struct {
		filename      string
		expectUpgrade UpgradeInfo
		expectErr     bool
	}{{
		filename:      "f1-good.json",
		expectUpgrade: UpgradeInfo{Name: "upgrade1", Info: "some info"},
		expectErr:     false,
	}, {
		filename:      "f2-bad-type.json",
		expectUpgrade: UpgradeInfo{},
		expectErr:     true,
	}, {
		filename:      "f3-empty.json",
		expectUpgrade: UpgradeInfo{},
		expectErr:     true,
	}, {
		filename:      "f4-empty-obj.json", // partial or empty objects also work!
		expectUpgrade: UpgradeInfo{},
		expectErr:     false,
	}, {
		filename:      "unknown.json",
		expectUpgrade: UpgradeInfo{},
		expectErr:     true,
	}}

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
