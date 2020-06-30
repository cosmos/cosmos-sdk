package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	// DefaultIBCVersion represents the latest supported version of IBC
	DefaultIBCVersion = "1.0.0"
)

// GetCompatibleVersions returns a descending ordered set of compatible IBC versions
// for the caller chain's connection end. The latest supported version should be
// first element and the set should decsend to the oldest supported version.
func GetCompatibleVersions() []string {
	return []string{DefaultIBCVersion}
}

// IsSupportedVersion returns true if the version provided is supported.
func IsSupportedVersion(version string, supportedVersions []string) bool {
	for _, supportedVer := range supportedVersions {
		if supportedVer == version {
			return true
		}
	}
	return false
}

// PickVersion iterates over the descending ordered set of compatible IBC versions
// and selects the first version that is supported by the counterparty.
func PickVersion(counterpartyVersions []string) (string, error) {
	versions := GetCompatibleVersions()

	for _, ver := range versions {
		if IsSupportedVersion(ver, counterpartyVersions) {
			return ver, nil
		}
	}

	return "", sdkerrors.Wrapf(
		ErrVersionNegotiationFailed,
		"failed to find a matching counterparty version (%s) from the supported version list (%s)", counterpartyVersions, versions,
	)
}
