package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
)

// testing of invalid version formats exist within 24-host/validate_test.go
func TestUnpackVersion(t *testing.T) {
	testCases := []struct {
		name          string
		version       string
		expIdentifier string
		expFeatures   []string
		expPass       bool
	}{
		{"valid version", "(1,[ORDERED channel,UNORDERED channel])", "1", []string{"ORDERED channel", "UNORDERED channel"}, true},
		{"valid empty features", "(1,[])", "1", []string{}, true},
		{"empty identifier", "(,[features])", "", []string{}, false},
		{"invalid version", "identifier,[features]", "", []string{}, false},
		{"empty string", "  ", "", []string{}, false},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			identifier, features, err := types.UnpackVersion(tc.version)

			if tc.expPass {
				require.NoError(t, err)
				require.Equal(t, tc.expIdentifier, identifier)
				require.Equal(t, tc.expFeatures, features)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestFindSupportedVersion(t *testing.T) {
	testCases := []struct {
		name              string
		version           string
		supportedVersions []string
		expVersion        string
		expFound          bool
	}{
		{"valid supported version", types.DefaultIBCVersion, types.GetCompatibleVersions(), types.DefaultIBCVersion, true},
		{"empty (invalid) version", "", types.GetCompatibleVersions(), "", false},
		{"empty supported versions", types.DefaultIBCVersion, []string{}, "", false},
		{"desired version is last", types.DefaultIBCVersion, []string{"(validversion,[])", "(2,[feature])", "(3,[])", types.DefaultIBCVersion}, types.DefaultIBCVersion, true},
		{"desired version identifier with different feature set", "(1,[features])", types.GetCompatibleVersions(), types.DefaultIBCVersion, true},
		{"version not supported", "(2,[DAG])", types.GetCompatibleVersions(), "", false},
	}

	for i, tc := range testCases {
		version, found := types.FindSupportedVersion(tc.version, tc.supportedVersions)

		require.Equal(t, tc.expVersion, version, "test case %d: %s", i, tc.name)
		require.Equal(t, tc.expFound, found, "test case %d: %s", i, tc.name)
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
		{"valid version in counterparty versions", []string{"(version1,[])", "(2.0.0,[DAG,ZK])", types.DefaultIBCVersion}, types.DefaultIBCVersion, true},
		{"valid identifier match but empty feature set", []string{"(1,[DAG,ORDERED-ZK,UNORDERED-zk])"}, "(1,[])", true},
		{"empty counterparty versions", []string{}, "", false},
		{"non-matching counterparty versions", []string{"(2.0.0,[])"}, "", false},
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

func TestVerifyProposedFeatureSet(t *testing.T) {
	testCases := []struct {
		name             string
		proposedVersion  string
		supportedVersion string
		expPass          bool
	}{
		{"entire feature set supported", types.DefaultIBCVersion, types.CreateVersionString("1", []string{"ORDER_ORDERED", "ORDER_UNORDERED", "ORDER_DAG"}), true},
		{"empty feature sets", types.CreateVersionString("1", []string{}), types.DefaultIBCVersion, true},
		{"one feature missing", types.DefaultIBCVersion, types.CreateVersionString("1", []string{"ORDER_UNORDERED", "ORDER_DAG"}), false},
		{"both features missing", types.DefaultIBCVersion, types.CreateVersionString("1", []string{"ORDER_DAG"}), false},
	}

	for i, tc := range testCases {
		supported := types.VerifyProposedFeatureSet(tc.proposedVersion, tc.supportedVersion)

		require.Equal(t, tc.expPass, supported, "test case %d: %s", i, tc.name)
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
