// Package version is a convenience utility that provides SDK
// consumers with a ready-to-use version command that
// produces apps versioning information based on flags
// passed at compile time.
//
// Configure the version command
//
// The version command can be just added to your cobra root command.
// At build time, the variables Name, Version, Commit, and BuildTags
// can be passed as build flags as shown in the following example:
//
//  go build -X github.com/cosmos/cosmos-sdk/version.Name=gaia \
//   -X github.com/cosmos/cosmos-sdk/version.ServerName=gaiad \
//   -X github.com/cosmos/cosmos-sdk/version.ClientName=gaiacli \
//   -X github.com/cosmos/cosmos-sdk/version.Version=1.0 \
//   -X github.com/cosmos/cosmos-sdk/version.Commit=f0f7b7dab7e36c20b757cebce0e8f4fc5b95de60 \
//   -X "github.com/cosmos/cosmos-sdk/version.BuildTags=linux darwin amd64"
package version

import (
	"fmt"
	"runtime"
)

var (
	// application's name
	Name = ""
	// server binary name
	ServerName = "<appd>"
	// client binary name
	ClientName = "<appcli>"
	// application's version string
	Version = ""
	// commit
	Commit = ""
	// hash of the go.sum file
	GoSumHash = ""
	// build tags
	BuildTags = ""
)

type versionInfo struct {
	Name       string `json:"name"`
	ServerName string `json:"server_name"`
	ClientName string `json:"client_name"`
	Version    string `json:"version"`
	GitCommit  string `json:"commit"`
	BuildTags  string `json:"build_tags"`
	GoVersion  string `json:"go"`
}

func (v versionInfo) String() string {
	return fmt.Sprintf(`%s: %s
git commit: %s
build tags: %s
%s`,
		v.Name, v.Version, v.GitCommit, v.BuildTags, v.GoVersion,
	)
}

func newVersionInfo() versionInfo {
	return versionInfo{
		Name:       Name,
		ServerName: ServerName,
		ClientName: ClientName,
		Version:    Version,
		GitCommit:  Commit,
		BuildTags:  BuildTags,
		GoVersion:  fmt.Sprintf("go version %s %s/%s", runtime.Version(), runtime.GOOS, runtime.GOARCH),
	}
}
