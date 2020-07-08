package types

var (
	supportedVersions []string

	allowNilFeatureSetMap map[string]bool
)

// RegisterVersions adds the supplied list of versions to the supportedVersions
// slice used in channel version negotiation. The provided versions are expected
// to be ordered in descending order of preference (most preferred version is
// first). This function is expected to be called by application modules
// during the `InitGenesis` function.
func RegisterVersions(versions ...string) {
	supportedVersions = append(supportedVersions, versions...)
}

// AllowNilFeatureSet registers a version to be compatible with nil feature
// set intersections agreed upon during a handshake negotiation. This function
// is expected to be called by application modules during the `InitGensis`
// function.
func AllowNilFeatureSet(version string) {
	allowNilFeatureSetMap[version] = true
}

// GetCompatibleVersions
func GetCompatibleVersions() []string {
	return supportedVersions
}

func GetAllowNilFeatureSetMap() map[string]bool {
	return allowNilFeatureSetMap
}
