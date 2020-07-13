package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
)

func TestValidateVersion(t *testing.T) {
	testCases := []struct {
		name    string
		version types.Version
		expPass bool
	}{
		{"valid version", types.DefaultIBCVersion, true},
		{"valid empty feature set", types.NewVersion(types.DefaultIBCVersionIdentifier, []string{}), true},
		{"empty version identifier", types.NewVersion("       ", []string{"ORDER_UNORDERED"}), false},
		{"empty feature", types.NewVersion(types.DefaultIBCVersionIdentifier, []string{"ORDER_UNORDERED", "   "}), false},
	}

	for i, tc := range testCases {
		encodedVersion, err := tc.version.ToString()
		require.NoError(t, err, "test case %d failed to marshal version string: %s", i, tc.name)

		err = types.ValidateVersion(encodedVersion)

		if tc.expPass {
			require.NoError(t, err, "valid test case %d failed: %s", i, tc.name)
		} else {
			require.Error(t, err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}

func TestFindSupportedVersion(t *testing.T) {
	testCases := []struct {
		name              string
		version           types.Version
		supportedVersions []types.Version
		expVersion        types.Version
		expFound          bool
	}{
		{"valid supported version", types.DefaultIBCVersion, types.GetCompatibleVersions(), types.DefaultIBCVersion, true},
		{"empty (invalid) version", types.Version{}, types.GetCompatibleVersions(), types.Version{}, false},
		{"empty supported versions", types.DefaultIBCVersion, []types.Version{}, types.Version{}, false},
		{"desired version is last", types.DefaultIBCVersion, []types.Version{types.NewVersion("1.1", nil), types.NewVersion("2", []string{"ORDER_UNORDERED"}), types.NewVersion("3", nil), types.DefaultIBCVersion}, types.DefaultIBCVersion, true},
		{"desired version identifier with different feature set", types.NewVersion(types.DefaultIBCVersionIdentifier, []string{"ORDER_DAG"}), types.GetCompatibleVersions(), types.DefaultIBCVersion, true},
		{"version not supported", types.NewVersion("2", []string{"ORDER_DAG"}), types.GetCompatibleVersions(), types.Version{}, false},
	}

	for i, tc := range testCases {
		version, found := types.FindSupportedVersion(tc.version, tc.supportedVersions)

		require.Equal(t, tc.expVersion.GetIdentifier(), version.GetIdentifier(), "test case %d: %s", i, tc.name)
		require.Equal(t, tc.expFound, found, "test case %d: %s", i, tc.name)
	}
}

func TestPickVersion(t *testing.T) {
	testCases := []struct {
		name                 string
		counterpartyVersions []types.Version
		expVer               types.Version
		expPass              bool
	}{
		{"valid default ibc version", types.GetCompatibleVersions(), types.DefaultIBCVersion, true},
		{"valid version in counterparty versions", []types.Version{types.NewVersion("version1", nil), types.NewVersion("2.0.0", []string{"ORDER_UNORDERED-ZK"}), types.DefaultIBCVersion}, types.DefaultIBCVersion, true},
		{"valid identifier match but empty feature set not allowed", []types.Version{types.NewVersion(types.DefaultIBCVersionIdentifier, []string{"DAG", "ORDERED-ZK", "UNORDERED-zk]"})}, types.NewVersion(types.DefaultIBCVersionIdentifier, nil), false},
		{"empty counterparty versions", []types.Version{}, types.Version{}, false},
		{"non-matching counterparty versions", []types.Version{types.NewVersion("2.0.0", nil)}, types.Version{}, false},
	}

	for i, tc := range testCases {
		encodedCounterpartyVersions, err := types.VersionsToStrings(tc.counterpartyVersions)
		require.NoError(t, err)

		encodedVersion, err := types.PickVersion(encodedCounterpartyVersions)

		if tc.expPass {
			require.NoError(t, err, "valid test case %d failed: %s", i, tc.name)

			version, err := types.StringToVersion(encodedVersion)
			require.NoError(t, err)
			require.Equal(t, tc.expVer, version, "valid test case %d falied: %s", i, tc.name)
		} else {
			require.Error(t, err, "invalid test case %d passed: %s", i, tc.name)
			require.Equal(t, "", encodedVersion, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}

func TestVerifyProposedVersion(t *testing.T) {
	testCases := []struct {
		name             string
		proposedVersion  types.Version
		supportedVersion types.Version
		expPass          bool
	}{
		{"entire feature set supported", types.DefaultIBCVersion, types.NewVersion("1", []string{"ORDER_ORDERED", "ORDER_UNORDERED", "ORDER_DAG"}), true},
		{"empty feature sets not supported", types.NewVersion("1", []string{}), types.DefaultIBCVersion, false},
		{"one feature missing", types.DefaultIBCVersion, types.NewVersion("1", []string{"ORDER_UNORDERED", "ORDER_DAG"}), false},
		{"both features missing", types.DefaultIBCVersion, types.NewVersion("1", []string{"ORDER_DAG"}), false},
	}

	for i, tc := range testCases {
		err := types.VerifyProposedVersion(tc.proposedVersion, tc.supportedVersion)

		if tc.expPass {
			require.NoError(t, err, "test case %d: %s", i, tc.name)
		} else {
			require.Error(t, err, "test case %d: %s", i, tc.name)
		}
	}

}

func TestVerifySupportedFeature(t *testing.T) {
	nilFeatures, err := types.NewVersion(types.DefaultIBCVersionIdentifier, nil).ToString()
	require.NoError(t, err)

	testCases := []struct {
		name    string
		version string
		feature string
		expPass bool
	}{
		{"check ORDERED supported", ibctesting.ConnectionVersion, "ORDER_ORDERED", true},
		{"check UNORDERED supported", ibctesting.ConnectionVersion, "ORDER_UNORDERED", true},
		{"check DAG unsupported", ibctesting.ConnectionVersion, "ORDER_DAG", false},
		{"check empty feature set returns false", nilFeatures, "ORDER_ORDERED", false},
	}

	for i, tc := range testCases {
		supported := types.VerifySupportedFeature(tc.version, tc.feature)

		require.Equal(t, tc.expPass, supported, "test case %d: %s", i, tc.name)
	}
}
