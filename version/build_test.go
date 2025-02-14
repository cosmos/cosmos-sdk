package version

import (
	"runtime/debug"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_getSDKBuildInfo(t *testing.T) {
	tests := []struct {
		name           string
		debugBuildInfo *debug.BuildInfo
		want           sdkBuildInfo
	}{
		{
			name: "no deps",
			debugBuildInfo: &debug.BuildInfo{
				Deps: nil,
			},
			want: sdkBuildInfo{},
		},
		{
			name: "cosmos-sdk dep only",
			debugBuildInfo: &debug.BuildInfo{
				Deps: []*debug.Module{
					{
						Path:    "github.com/cosmos/cosmos-sdk",
						Version: "v2.0.0",
					},
				},
			},
			want: sdkBuildInfo{
				sdkVersion: "v2.0.0",
			},
		},
		{
			name: "all depo",
			debugBuildInfo: &debug.BuildInfo{
				Deps: []*debug.Module{
					{
						Path:    "github.com/cosmos/cosmos-sdk",
						Version: "v2.0.0",
					},
					{
						Path:    "cosmossdk.io/server/v2/cometbft",
						Version: "v2.0.1",
					},
					{
						Path:    "cosmossdk.io/runtime/v2",
						Version: "v2.0.2",
					},
					{
						Path:    "cosmossdk.io/server/v2/stf",
						Version: "v2.0.3",
					},
				},
			},
			want: sdkBuildInfo{
				sdkVersion:         "v2.0.0",
				cometServerVersion: "v2.0.1",
				runtimeVersion:     "v2.0.2",
				stfVersion:         "v2.0.3",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, getSDKBuildInfo(tt.debugBuildInfo))
		})
	}
}

func Test_extractVersionFromBuildInfo(t *testing.T) {
	tests := []struct {
		name string
		dep  *debug.Module
		want string
	}{
		{
			name: "no replace",
			dep: &debug.Module{
				Path:    "github.com/cosmos/cosmos-sdk",
				Version: "v2.0.0",
			},
			want: "v2.0.0",
		},
		{
			name: "devel replace ",
			dep: &debug.Module{
				Path:    "github.com/cosmos/cosmos-sdk",
				Version: "v2.0.0",
				Replace: &debug.Module{
					Version: "(devel)",
				},
			},
			want: "v2.0.0",
		},
		{
			name: "non-devel replace ",
			dep: &debug.Module{
				Path:    "github.com/cosmos/cosmos-sdk",
				Version: "v2.0.0",
				Replace: &debug.Module{
					Version: "v1.0.3",
				},
			},
			want: "v1.0.3",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, extractVersionFromBuildInfo(tt.dep))
		})
	}
}

func Test_depsFromBuildInfo(t *testing.T) {
	modules := []*debug.Module{
		{
			Path:    "github.com/cosmos/cosmos-sdk",
			Version: "v2.0.0",
		},
		{
			Path:    "cosmossdk.io/server/v2/cometbft",
			Version: "v2.0.1",
		},
		{
			Path:    "cosmossdk.io/runtime/v2",
			Version: "v2.0.2",
		},
		{
			Path:    "cosmossdk.io/server/v2/stf",
			Version: "v2.0.3",
		},
	}

	tests := []struct {
		name           string
		debugBuildInfo *debug.BuildInfo
		want           []buildDep
	}{
		{
			name:           "no deps",
			debugBuildInfo: &debug.BuildInfo{},
			want:           nil,
		},
		{
			name: "deps",
			debugBuildInfo: &debug.BuildInfo{
				Deps: modules,
			},
			want: []buildDep{
				{modules[0]},
				{modules[1]},
				{modules[2]},
				{modules[3]},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, depsFromBuildInfo(tt.debugBuildInfo))
		})
	}
}
