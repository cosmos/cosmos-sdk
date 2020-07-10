package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
)

func TestValidateBasic(t *Testing.T) {
	testCases := []struct {
		name    string
		version types.Version
		expPass bool
	}{
		{"valid version", types.DefaultIBCVersion, true},
		{"valid empty feature set", types.NewVersion(types.DefaultIBCVersionIdentifier, []string{})},
		{"empty version identifier", types.NewVersion("       ", []string{"ORDER_UNORDERED", false})},
		{"empty feature", types.NewVersion(types.DefaultIBCVersionIdentifier, []string{"ORDER_UNORDERED", "   "})},
	}

	for i, tc := range testCases {

		err := tc.version.ValidateBasic()

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
		version           Version
		supportedVersions []Version
		expVersion        Version
		expFound          bool
	}{
		{"valid supported version", types.DefaultIBCVersion, types.GetCompatibleVersions(), types.DefaultIBCVersion, true},
		{"empty (invalid) version", "", types.GetCompatibleVersions(), nil, false},
		{"empty supported versions", types.DefaultIBCVersion, []Version{}, nil, false},
		{"desired version is last", types.DefaultIBCVersion, []Version{types.NewVersion("1.1", string{}), types.NewVersion("2", []string{"ORDER_UNORDERED"}), types.NewVersion("3", nil), types.DefaultIBCVersion}, types.DefaultIBCVersion, true},
		{"desired version identifier with different feature set", types.NewVersion(DefaultIBCVersionIdentifier, []string{"ORDER_DAG"}), types.GetCompatibleVersions(), types.DefaultIBCVersion, true},
		{"version not supported", types.NewVersion("2", []string{"ORDER_DAG"}), types.GetCompatibleVersions(), nil, false},
	}

	for i, tc := range testCases {
		version, found := types.FindSupportedVersion(tc.version, tc.supportedVersions)

		require.Equal(t, tc.expVersion.Identifer, version.Identifier, "test case %d: %s", i, tc.name)
		require.Equal(t, tc.expFound, found, "test case %d: %s", i, tc.name)
	}
}

func TestPickVersion(t *testing.T) {
	testCases := []struct {
		name                 string
		counterpartyVersions []Version
		expVer               Version
		expPass              bool
	}{
		{"valid default ibc version", types.GetCompatibleVersions(), types.DefaultIBCVersion, true},
		{"valid version in counterparty versions", []Version{types.NewVersion("version1", nil), types.NewVersion("2.0.0", []string("ORDER_UNORDERED-ZK")), types.DefaultIBCVersion}, types.DefaultIBCVersion, true},
		{"valid identifier match but empty feature set not allowed", []Version{types.NewVersion(types.DefaultIBCVersionIdentifier, []string{"DAG", "ORDERED-ZK", "UNORDERED-zk]"}), types.NewVersion(types.DefaultIBCVersionIdentifier, nil)}, false},
		{"empty counterparty versions", []Version{}, nil, false},
		{"non-matching counterparty versions", []Version{types.NewVersion("2.0.0", nil)}, nil, false},
	}

	for i, tc := range testCases {
		version, err := types.PickVersion(tc.counterpartyVersions)

		if tc.expPass {
			require.NoError(t, err, "valid test case %d failed: %s", i, tc.name)
			require.Equal(t, tc.expVer, version, "valid test case %d falied: %s", i, tc.name)
		} else {
			require.Error(t, err, "invalid test case %d passed: %s", i, tc.name)
			require.Equal(t, nil, version, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}

func TestVerifyProposedVersion(t *testing.T) {
	testCases := []struct {
		name             string
		proposedVersion  string
		supportedVersion string
		expPass          bool
	}{
		{"entire feature set supported", types.DefaultIBCVersion, types.CreateVersionString("1", []string{"ORDER_ORDERED", "ORDER_UNORDERED", "ORDER_DAG"}), true},
		{"empty feature sets not supported", types.CreateVersionString("1", []string{}), types.DefaultIBCVersion, false},
		{"one feature missing", types.DefaultIBCVersion, types.CreateVersionString("1", []string{"ORDER_UNORDERED", "ORDER_DAG"}), false},
		{"both features missing", types.DefaultIBCVersion, types.CreateVersionString("1", []string{"ORDER_DAG"}), false},
		{"could not unpack proposed version", "(invalid version)", types.DefaultIBCVersion, false},
		{"could not unpack supported version", types.DefaultIBCVersion, "(invalid version)", false},
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
	testCases := []struct {
		name    string
		version string
		feature string
		expPass bool
	}{
		{"check ORDERED supported", types.DefaultIBCVersion, "ORDER_ORDERED", true},
		{"check UNORDERED supported", types.DefaultIBCVersion, "ORDER_UNORDERED", true},
		{"check DAG unsupported", types.DefaultIBCVersion, "ORDER_DAG", false},
		{"check empty feature set returns false", types.CreateVersionString("1", []string{}), "ORDER_ORDERED", false},
	}

	for i, tc := range testCases {
		supported := types.VerifySupportedFeature(tc.version, tc.feature)

		require.Equal(t, tc.expPass, supported, "test case %d: %s", i, tc.name)
	}
}
