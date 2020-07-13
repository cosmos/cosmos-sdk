package types

import (
	"strings"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
)

var (
	// DefaultIBCVersion represents the latest supported version of IBC used
	// in connection version negotiation. The current version supports only
	// ORDERED and UNORDERED channels and requires at least one channel type
	// to be agreed upon.
	DefaultIBCVersion = NewVersion(DefaultIBCVersionIdentifier, []string{"ORDER_ORDERED", "ORDER_UNORDERED"})

	// DefaultIBCVersionIdentifier is the IBC v1.0.0 protocol version identifier
	DefaultIBCVersionIdentifier = "1"

	// AllowNilFeatureSet is a helper map to indicate if a specified version
	// identifier is allowed to have a nil feature set. Any versions supported,
	// but not included in the map default to not supporting nil feature sets.
	allowNilFeatureSet = map[string]bool{
		DefaultIBCVersionIdentifier: false,
	}
)

// NewVersion returns a new instance of Version.
func NewVersion(identifier string, features []string) Version {
	return Version{
		Identifier: identifier,
		Features:   features,
	}
}

// GetIdentifier implements the VersionI interface
func (version Version) GetIdentifier() string {
	return version.Identifier
}

// GetFeatures implements the VersionI interface
func (version Version) GetFeatures() []string {
	return version.Features
}

// ValidateVersion does basic validation of the version identifier and
// features. It unmarshals the version string into a Version object.
func ValidateVersion(ver string) error {
	var version Version
	if err := SubModuleCdc.UnmarshalBinaryBare([]byte(ver), &version); err != nil {
		return sdkerrors.Wrap(err, "failed to unmarshal version string %s", ver)
	}

	if strings.TrimSpace(version.Identifier) == "" {
		return sdkerrors.Wrap(ErrInvalidVersion, "version identifier cannot be blank")
	}
	for i, feature := range version.Features {
		if strings.TrimSpace(feature) == "" {
			return sdkerrors.Wrapf(ErrInvalidVersion, "feature cannot be blank, index %d", i)
		}
	}

	return nil
}

// GetCompatibleVersions returns a descending ordered set of compatible IBC
// versions for the caller chain's connection end. The latest supported
// version should be first element and the set should descend to the oldest
// supported version.
func GetCompatibleVersions() []Version {
	return []Version{DefaultIBCVersion}
}

// FindSupportedVersion returns the version with a matching version identifier
// if it exists. The returned boolean is true if the version is found and
// false otherwise.
func FindSupportedVersion(version Version, supportedVersions []Version) (Version, bool) {
	for _, supportedVersion := range supportedVersions {
		if version.GetIdentifier() == supportedVersion.GetIdentifier() {
			return supportedVersion, true
		}
	}
	return Version{}, false
}

// PickVersion iterates over the descending ordered set of compatible IBC
// versions and selects the first version with a version identifier that is
// supported by the counterparty. The returned version contains a feature
// set with the intersection of the features supported by the source and
// counterparty chains. If the feature set intersection is nil and this is
// not allowed for the choosen version identifier then the search for a
// compatible version continues. This function is called in the ConnOpenTry
// handshake procedure.
func PickVersion(counterpartyVersions []Version) (Version, error) {
	supportedVersions := GetCompatibleVersions()

	for _, supportedVersion := range supportedVersions {
		// check if the source version is supported by the counterparty
		if counterpartyVersion, found := FindSupportedVersion(supportedVersion, counterpartyVersions); found {

			featureSet := GetFeatureSetIntersection(supportedVersion.GetFeatures(), counterpartyVersion.GetFeatures())
			if len(featureSet) == 0 && !allowNilFeatureSet[supportedVersion.GetIdentifier()] {
				continue
			}

			return NewVersion(supportedVersion.GetIdentifier(), featureSet), nil
		}
	}

	return Version{}, sdkerrors.Wrapf(
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
// identifier.
func VerifyProposedVersion(proposedVersion, supportedVersion Version) error {
	// sanity check
	if proposedVersion.GetIdentifier() != supportedVersion.GetIdentifier() {
		return sdkerrors.Wrapf(
			ErrVersionNegotiationFailed,
			"proposed version identifier does not equal supported version identifier (%s != %s)", proposedVersion.GetIdentifier(), supportedVersion.GetIdentifier(),
		)
	}

	if len(proposedVersion.GetFeatures()) == 0 && !allowNilFeatureSet[proposedVersion.GetIdentifier()] {
		return sdkerrors.Wrapf(
			ErrVersionNegotiationFailed,
			"nil feature sets are not supported for version identifier (%s)", proposedVersion.GetIdentifier(),
		)
	}

	for _, proposedFeature := range proposedVersion.GetFeatures() {
		if !contains(proposedFeature, supportedVersion.GetFeatures()) {
			return sdkerrors.Wrapf(
				ErrVersionNegotiationFailed,
				"proposed feature (%s) is not a supported feature set (%s)", proposedFeature, supportedVersion.GetFeatures(),
			)
		}
	}

	return nil
}

// VerifySupportedFeature takes in a version and feature string and returns
// true if the feature is supported by the version and false otherwise.
func VerifySupportedFeature(version Version, feature string) bool {
	for _, f := range version.GetFeatures() {
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
