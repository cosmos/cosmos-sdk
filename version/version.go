// Package version is a convenience utility that provides SDK
// consumers with a ready-to-use version command that
// produces apps versioning information based on flags
// passed at compile time.
//
// # Configure the version command
//
// The version command can be just added to your cobra root command.
// At build time, the variables Name, Version, Commit, and BuildTags
// can be passed as build flags as shown in the following example:
//
//	go build -X github.com/cosmos/cosmos-sdk/version.Name=gaia \
//	 -X github.com/cosmos/cosmos-sdk/version.AppName=gaiad \
//	 -X github.com/cosmos/cosmos-sdk/version.Version=1.0 \
//	 -X github.com/cosmos/cosmos-sdk/version.Commit=f0f7b7dab7e36c20b757cebce0e8f4fc5b95de60 \
//	 -X "github.com/cosmos/cosmos-sdk/version.BuildTags=linux darwin amd64"
package version

import (
	"encoding/json"
	"fmt"
	"runtime"
	"runtime/debug"
)

// ContextKey is used to store the ExtraInfo in the context.
type ContextKey struct{}

var (
	// application's name
	Name = ""
	// application binary name
	AppName = "<appd>"
	// application's version string
	Version = ""
	// commit
	Commit = ""
	// build tags
	BuildTags = ""
)

type sdkBuildInfo struct {
	sdkVersion         string
	runtimeVersion     string
	stfVersion         string
	cometServerVersion string
}

func getSDKBuildInfo(debugBuildInfo *debug.BuildInfo) sdkBuildInfo {
	var buildInfo sdkBuildInfo
	for _, dep := range debugBuildInfo.Deps {
		switch dep.Path {
		case "github.com/cosmos/cosmos-sdk":
			buildInfo.sdkVersion = extractVersionFromBuildInfo(dep)
		case "cosmossdk.io/server/v2/cometbft":
			buildInfo.cometServerVersion = extractVersionFromBuildInfo(dep)
		case "cosmossdk.io/runtime/v2":
			buildInfo.runtimeVersion = extractVersionFromBuildInfo(dep)
		case "cosmossdk.io/server/v2/stf":
			buildInfo.stfVersion = extractVersionFromBuildInfo(dep)
		}
	}

	return buildInfo
}

func extractVersionFromBuildInfo(dep *debug.Module) string {
	if dep.Replace != nil && dep.Replace.Version != "(devel)" {
		return dep.Replace.Version
	}

	return dep.Version
}

// ExtraInfo contains a set of extra information provided by apps
type ExtraInfo map[string]string

// Info defines the application version information.
type Info struct {
	Name               string     `json:"name" yaml:"name"`
	AppName            string     `json:"server_name" yaml:"server_name"`
	Version            string     `json:"version" yaml:"version"`
	GitCommit          string     `json:"commit" yaml:"commit"`
	BuildTags          string     `json:"build_tags" yaml:"build_tags"`
	GoVersion          string     `json:"go" yaml:"go"`
	BuildDeps          []buildDep `json:"build_deps" yaml:"build_deps"`
	CosmosSdkVersion   string     `json:"cosmos_sdk_version" yaml:"cosmos_sdk_version"`
	RuntimeVersion     string     `json:"runtime_version,omitempty" yaml:"runtime_version,omitempty"`
	StfVersion         string     `json:"stf_version,omitempty" yaml:"stf_version,omitempty"`
	CometServerVersion string     `json:"comet_server_version,omitempty" yaml:"comet_server_version,omitempty"`
	ExtraInfo          ExtraInfo  `json:"extra_info,omitempty" yaml:"extra_info,omitempty"`
}

func NewInfo() Info {
	info := Info{
		Name:             Name,
		AppName:          AppName,
		Version:          Version,
		GitCommit:        Commit,
		BuildTags:        BuildTags,
		GoVersion:        fmt.Sprintf("go version %s %s/%s", runtime.Version(), runtime.GOOS, runtime.GOARCH),
		CosmosSdkVersion: "unable to read deps",
	}

	// use debug info more granular build info if available
	debugBuildInfo, ok := debug.ReadBuildInfo()
	if ok {
		info.BuildDeps = depsFromBuildInfo(debugBuildInfo)
		sdkBuildInfo := getSDKBuildInfo(debugBuildInfo)
		info.CosmosSdkVersion = sdkBuildInfo.sdkVersion
		info.RuntimeVersion = sdkBuildInfo.runtimeVersion
		info.StfVersion = sdkBuildInfo.stfVersion
		info.CometServerVersion = sdkBuildInfo.cometServerVersion
	}

	return info
}

func (vi Info) String() string {
	return fmt.Sprintf(`%s: %s
git commit: %s
build tags: %s
%s`,
		vi.Name, vi.Version, vi.GitCommit, vi.BuildTags, vi.GoVersion,
	)
}

func depsFromBuildInfo(debugBuildInfo *debug.BuildInfo) (deps []buildDep) {
	for _, dep := range debugBuildInfo.Deps {
		deps = append(deps, buildDep{dep})
	}

	return
}

type buildDep struct {
	*debug.Module
}

func (d buildDep) String() string {
	if d.Replace != nil {
		return fmt.Sprintf("%s@%s => %s@%s", d.Path, d.Version, d.Replace.Path, d.Replace.Version)
	}

	return fmt.Sprintf("%s@%s", d.Path, d.Version)
}

func (d buildDep) MarshalJSON() ([]byte, error)      { return json.Marshal(d.String()) }
func (d buildDep) MarshalYAML() (interface{}, error) { return d.String(), nil }
