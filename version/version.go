//nolint
package version

import (
	"fmt"
	"runtime"
)

// Variables set by build flags
var (
	Commit    = ""
	Version   = ""
	GoModHash = ""
	BuildTags = ""
)

type versionInfo struct {
	CosmosSDK string `json:"cosmos_sdk"`
	GitCommit string `json:"commit"`
	GoSumHash string `json:"gosum_hash"`
	BuildTags string `json:"build_tags"`
	GoVersion string `json:"go"`
}

func (v versionInfo) String() string {
	return fmt.Sprintf(`cosmos-sdk: %s
git commit: %s
gomod hash: %s
build tags: %s
%s`, v.CosmosSDK, v.GitCommit, v.GoModHash, v.BuildTags, v.GoVersion)
}

func newVersionInfo() versionInfo {
	return versionInfo{
		Version,
		Commit,
		GoModHash,
		BuildTags,
		fmt.Sprintf("go version %s %s/%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)}
}
