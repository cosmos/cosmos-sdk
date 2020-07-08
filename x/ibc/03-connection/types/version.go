package types

import (
	"fmt"
	"strings"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

var (
	// DefaultIBCVersion represents the latest supported version of IBC used
	// in connection version negotiation. The current version supports only
	// ORDERED and UNORDERED channels and requires at least one channel type
	// to be agreed upon.
	DefaultIBCVersion           = CreateVersionString(DefaultIBCVersionIdentifier, []string{"ORDER_ORDERED", "ORDER_UNORDERED"})
	DefaultIBCVersionIdentifier = "1"

	// AllowNilFeatureSet is a helper map to indicate if a specified version
	// identifier is allowed to have a nil feature set. Any versions supported,
	// but not included in the map default to not supporting nil feature sets.
	allowNilFeatureSet = map[string]bool{
		DefaultIBCVersionIdentifier: false,
	}
)

// GetCompatibleVersions returns a descending ordered set of compatible IBC
// versions for the caller chain's connection end. The latest supported
// version should be first element and the set should descend to the oldest
// supported version.
func GetCompatibleVersions() []string {
	return []string{DefaultIBCVersion}
}

// CreateVersionString constructs a valid connection version given a
// version identifier and feature set. The format is written as:
//
// ([version_identifier],[feature_0,feature_1,feature_2...])
//
// A connection version is considered invalid if it is not in this format
// or violates one of these rules:
// - the version identifier is empty or contains commas
// - a specified feature contains commas
func CreateVersionString(identifier string, featureSet []string) string {
	return fmt.Sprintf("(%s,[%s])", identifier, strings.Join(featureSet, ","))
}

// UnpackVersion parses a version string and returns the identifier and the
// feature set of a version. An error is returned if the version is not valid.
func UnpackVersion(version string) (string, []string, error) {
	// validate version so valid splitting assumptions can be made
	if err := host.VersionValidator(version); err != nil {
		return "", nil, err
	}

	// peel off prefix and suffix of the tuple
	version = strings.TrimPrefix(version, "(")
	version = strings.TrimSuffix(version, ")")

	// split into identifier and feature set
	splitVersion := strings.SplitN(version, ",", 2)
	if splitVersion[0] == version {
		return "", nil, sdkerrors.Wrapf(ErrInvalidVersion, "failed to split version '%s' into identifier and features", version)
	}
	identifier := splitVersion[0]

	// peel off prefix and suffix of features
	featureSet := strings.TrimPrefix(splitVersion[1], "[")
	featureSet = strings.TrimSuffix(featureSet, "]")

	// check if features are empty
	if len(featureSet) == 0 {
		return identifier, []string{}, nil
	}

	features := strings.Split(featureSet, ",")

	return identifier, features, nil
}

// FindSupportedVersion returns the version with a matching version identifier
// if it exists. The returned boolean is true if the version is found and
// false otherwise.
func FindSupportedVersion(version string, supportedVersions []string) (string, bool) {
	identifier, _, err := UnpackVersion(version)
	if err != nil {
		return "", false
	}

	for _, supportedVersion := range supportedVersions {
		supportedIdentifier, _, err := UnpackVersion(supportedVersion)
		if err != nil {
			continue
		}

		if identifier == supportedIdentifier {
			return supportedVersion, true
		}
	}
	return "", false
}

// PickVersion iterates over the descending ordered set of compatible IBC
// versions and selects the first version with a version identifier that is
// supported by the counterparty. The returned version contains a feature
// set with the intersection of the features supported by the source and
// counterparty chains. This function is called in the ConnOpenTry handshake
// procedure.
func PickVersion(counterpartyVersions []string) (string, error) {
	supportedVersions := GetCompatibleVersions()

	for _, ver := range supportedVersions {
		// check if the source version is supported by the counterparty
		if counterpartyVer, found := FindSupportedVersion(ver, counterpartyVersions); found {
			sourceIdentifier, sourceFeatures, err := UnpackVersion(ver)
			if err != nil {
				return "", err
			}

			_, counterpartyFeatures, err := UnpackVersion(counterpartyVer)
			if err != nil {
				return "", err
			}

			featureSet := GetFeatureSetIntersection(sourceFeatures, counterpartyFeatures)
			if len(featureSet) == 0 && !allowNilFeatureSet[ver] {
				continue
			}

			version := CreateVersionString(sourceIdentifier, featureSet)
			return version, nil
		}
	}

	return "", sdkerrors.Wrapf(
		ErrVersionNegotiationFailed,
		"failed to find a matching counterparty version (%s) from the supported version list (%s)", counterpartyVersions, supportedVersions,
	)
}

// GetFeatureSetIntersection returns the intersections of source feature set
// and the counterparty feature set. This is done by iterating over all the
// features in the source version and seeing if they exist in the feature
// set for the counterparty version.
func GetFeatureSetIntersection(sourceFeatureSet, counterpartyFeatureSet []string) (featureSet []string) {
	for _, feature := range sourceFeatureSet {
		if contains(feature, counterpartyFeatureSet) {
			featureSet = append(featureSet, feature)
		}
	}

	return featureSet
}

// VerifyProposedVersion verifies that the entire feature set in the
// proposed version is supported by this chain. If the feature set is
// empty it verifies that this is allowed for the specified version
// identifier. It also ensures that the supported version identifier
// matches the proposed version identifier.
func VerifyProposedVersion(proposedVersion, supportedVersion string) error {
	proposedIdentifier, proposedFeatureSet, err := UnpackVersion(proposedVersion)
	if err != nil {
		return sdkerrors.Wrap(err, "could not unpack proposed version")
	}

	if len(proposedFeatureSet) == 0 && !allowNilFeatureSet[proposedIdentifier] {
		return sdkerrors.Wrapf(
			ErrVersionNegotiationFailed,
			"nil feature sets are not supported for version identifier (%s)", proposedIdentifier,
		)
	}

	_, supportedFeatureSet, err := UnpackVersion(supportedVersion)
	if err != nil {
		return sdkerrors.Wrap(err, "could not unpack supported version")
	}

	for _, proposedFeature := range proposedFeatureSet {
		if !contains(proposedFeature, supportedFeatureSet) {
			return sdkerrors.Wrapf(
				ErrVersionNegotiationFailed,
				"proposed feature set (%s) is not a supported feature set (%s)", proposedFeatureSet, supportedFeatureSet,
			)
		}
	}

	return nil
}

// VerifySupportedFeature takes in a version string and feature string and returns
// true if the feature is supported by the version and false otherwise.
func VerifySupportedFeature(version, feature string) bool {
	_, featureSet, err := UnpackVersion(version)
	if err != nil {
		return false
	}

	for _, f := range featureSet {
		if f == feature {
			return true
		}
	}
	return false
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
