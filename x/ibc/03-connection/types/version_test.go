package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
)

func TestIsSupportedVersion(t *testing.T) {
	testCases := []struct {
		name              string
		version           string
		supportedVersions []string
		expRes            bool
	}{
		{"valid supported version", types.DefaultIBCVersion, types.GetCompatibleVersions(), true},
		{"empty version", "", types.GetCompatibleVersions(), false},
		{"empty supported versions", types.DefaultIBCVersion, []string{}, false},
		{"desired version is last", types.DefaultIBCVersion, []string{"1", "2  ", "  3", types.DefaultIBCVersion}, true},
		{"version not supported", "2.0.0", types.GetCompatibleVersions(), false},
	}

	for i, tc := range testCases {
		res := types.IsSupportedVersion(tc.version, tc.supportedVersions)

		require.Equal(t, tc.expRes, res, "test case %d: %s", i, tc.name)
	}
}

func TestPickVersion(t *testing.T) {
	testCases := []struct {
		name                 string
		counterpartyVersions []string
		expVer               string
		expPass              bool
	}{
		{"valid default ibc version", types.GetCompatibleVersions(), types.DefaultIBCVersion, true},
		{"valid version in counterparty versions", []string{"1", "2.0.0", types.DefaultIBCVersion}, types.DefaultIBCVersion, true},
		{"empty counterparty versions", []string{}, "", false},
		{"non-matching counterparty versions", []string{"2.0.0"}, "", false},
	}

	for i, tc := range testCases {
		version, err := types.PickVersion(tc.counterpartyVersions)

		if tc.expPass {
			require.NoError(t, err, "valid test case %d failed: %s", i, tc.name)
			require.Equal(t, tc.expVer, version, "valid test case %d falied: %s", i, tc.name)
		} else {
			require.Error(t, err, "invalid test case %d passed: %s", i, tc.name)
			require.Equal(t, "", version, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}
