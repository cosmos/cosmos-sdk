// This package is a convenience utility that provides SDK
// consumers with a ready-to-use version command that
// produces apps versioning information based on flags
// passed at compile time.
//
// Configure the version command
//
// The version command can be just added to your cobra root command.
// At build time, the variables Name, Version, Commit, GoSumHash, and
// BuildTags can be passed as build flags as shown in the following
// example:
//
//  go build -X github.com/cosmos/cosmos-sdk/version.Name=dapp \
//   -X github.com/cosmos/cosmos-sdk/version.Version=1.0 \
//   -X github.com/cosmos/cosmos-sdk/version.Commit=f0f7b7dab7e36c20b757cebce0e8f4fc5b95de60 \
//   -X "github.com/cosmos/cosmos-sdk/version.BuildTags=linux darwin amd64"
package version

import (
	"fmt"
	"runtime"
)

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
		fmt.Sprintf("go version %s %s/%s", runtime.Version(), runtime.GOOS, runtime.GOARCH)}
}
