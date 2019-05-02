//nolint
package version

import (
	"fmt"
	"runtime"
)

// Variables representin application's versioning
// information set at build time.
var (
	// Application's name
	Name = ""
	// Application's version string
	Version = ""
	// Commit
	Commit = ""
	// Hash of the go.sum file
	GoSumHash = ""
	// Build tags
	BuildTags = ""
)

type versionInfo struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	GitCommit string `json:"commit"`
	GoSumHash string `json:"gosum_hash"`
	BuildTags string `json:"build_tags"`
	GoVersion string `json:"go"`
}

func (v versionInfo) String() string {
	return fmt.Sprintf(`%s: %s
git commit: %s
go.sum hash: %s
build tags: %s
%s`, v.Name, v.Version, v.GitCommit, v.GoSumHash, v.BuildTags, v.GoVersion)
}

func newVersionInfo() versionInfo {
	return versionInfo{
		Name,
		Version,
		Commit,
		GoSumHash,
		BuildTags,
		fmt.Sprintf("go version %s %s/%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)}
}
