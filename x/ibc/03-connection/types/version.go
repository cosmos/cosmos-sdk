package types

import (
	"fmt"
	"strings"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	// DefaultIBCVersion represents the latest supported version of IBC.
	// The current version supports only ORDERED and UNORDERED channels.
	DefaultIBCVersion = CreateVersionString("1", []string{"ORDERED channel", "UNORDERED channel"})
)

// CreateVersionString constructs a valid connection version given a
// version number and feature set. The format is written as:
// [version-number]-[feature_0,feature_1,feature_2...]
func CreateVersionString(versionNumber string, featureSet []string) string {
	version := versionNumber
	for i, feature := range featureSet {
		if i == 0 {
			version = fmt.Sprintf("%s-%s", version, feature)
		} else {
			version = fmt.Sprintf("%s,%s", version, feature)
		}
	}

	return version
}

// GetVersionNumber returns the version number of a given connection version.
func GetVersionNumber(version string) (string, error) {
	splitVersion := strings.Split(version, "-")

	// check if no dash exists in the version
	if len(splitVersion) < 2 {
		return "", sdkerrors.Wrapf(
			ErrInvalidVersion,
			"could not retrieve the version number for version (%s)", version,
		)
	}

	return strings.TrimSpace(splitVersion[0]), nil
}

// GetFeatureSet returns the feature set for a given connection version.
func GetFeatureSet(version string) ([]string, error) {
	// only split version number and feature set
	splitVersion := strings.SplitN(version, "-", 1)
	if len(splitVersion) < 2 {
		return nil, sdkerrors.Wrapf(
			ErrInvalidVersion,
			"failed to retrieve the feature set for version (%s)", version,
		)
	}

	// split feature set
	return strings.Split(splitVersion[1], ","), nil
}

// GetCompatibleVersions returns a descending ordered set of compatible IBC
// versions for the caller chain's connection end. The latest supported
// version should be first element and the set should descend to the oldest
// supported version.
func GetCompatibleVersions() []string {
	return []string{DefaultIBCVersion}
}

// FindSupportedVersion returns the version with a matching version number
// if it exists. The returned boolean is true if the version is found and
// false otherwise.
func FindSupportedVersion(version string, supportedVersions []string) (string, bool) {
	versionNumber, err := GetVersionNumber(version)
	if err != nil {
		return "", false
	}

	for _, supportedVersion := range supportedVersions {
		supportedVersionNumber, err := GetVersionNumber(supportedVersion)
		if err != nil {
			continue
		}

		if supportedVersionNumber == versionNumber {
			return supportedVersion, true
		}
	}
	return "", false
}

// PickVersion iterates over the descending ordered set of compatible IBC
// versions and selects the first version that is supported by the counterparty.
func PickVersion(counterpartyVersions []string) (string, error) {
	versions := GetCompatibleVersions()

	for _, ver := range versions {
		if counterpartyVer, found := FindSupportedVersion(ver, counterpartyVersions); found {
			sourceFeatureSet, err := GetFeatureSet(ver)
			if err != nil {
				return "", err
			}

			counterpartyFeatureSet, err := GetFeatureSet(counterpartyVer)
			if err != nil {
				return "", err
			}

			featureSet := GetFeatureSetIntersection(sourceFeatureSet, counterpartyFeatureSet)

			versionNumber, err := GetVersionNumber(ver)
			if err != nil {
				return "", err
			}

			version := CreateVersionString(versionNumber, featureSet)
			return version, nil
		}
	}

	return "", sdkerrors.Wrapf(
		ErrVersionNegotiationFailed,
		"failed to find a matching counterparty version (%s) from the supported version list (%s)", counterpartyVersions, versions,
	)
}

// GetFeatureSetIntersection returns the intersections of feature set A and
// feature set B. This is done by iterating over all the features in A and
// seeing if they exist in the feature set for B.
func GetFeatureSetIntersection(featureSetA, featureSetB []string) (featureSet []string) {
	for _, feature := range featureSetA {
		if contains(feature, featureSetB) {
			featureSet = append(featureSet, feature)
		}
	}

	return featureSet
}

// VerifyProposedFeatureSet verifies that the entire feature set in the
// proposed version is supported.
func VerifyProposedFeatureSet(proposedVersion, supportedVersion string) bool {
	proposedFeatureSet, err := GetFeatureSet(proposedVersion)
	if err != nil {
		return false
	}

	supportedFeatureSet, err := GetFeatureSet(supportedVersion)
	if err != nil {
		return false
	}

	for _, proposedFeature := range proposedFeatureSet {
		if !contains(proposedFeature, supportedFeatureSet) {
			return false
		}
	}

	return true
}

// contains returns true if the provided string element exists within the
// string set.
func contains(elem string, set []string) bool {
	for _, element := range set {
		if strings.TrimSpace(elem) == strings.TrimSpace(element) {
			return true
		}
	}

	return false
}
