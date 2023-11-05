package util

import (
	"runtime/debug"
	"testing"
)

func TestIsSupportedVersion(t *testing.T) {
	mockBuildInfo := &debug.BuildInfo{
		Deps: []*debug.Module{
			{
				Path:    "github.com/cosmos/cosmos-sdk",
				Version: "v0.50.0",
			},
			{
				Path:    "cosmossdk.io/feegrant",
				Version: "v0.1.0",
			},
		},
	}

	cases := []struct {
		input    string
		expected bool
	}{
		{
			input:    "",
			expected: true,
		},
		{
			input:    "not a since comment",
			expected: true,
		},
		{
			input:    "// Since: cosmos-sdk v0.47",
			expected: true,
		},
		{
			input:    "// since: Cosmos-SDK 0.50",
			expected: true,
		},
		{
			input:    "// Since: cosmos-sdk v0.51",
			expected: false,
		},
		{
			input:    "// Since: cosmos-sdk v1.0.0",
			expected: false,
		},
		{
			input:    "// since: x/feegrant v0.1.0",
			expected: true,
		},
		{
			input:    "// since: feegrant v0.0.1",
			expected: true,
		},
		{
			input:    "// since: feegrant v0.1.0",
			expected: true,
		},
		{
			input:    "// since: feegrant v0.1",
			expected: true,
		},
		{
			input:    "// since: feegrant v0.1.1",
			expected: false,
		},
		{
			input:    "// since: feegrant v0.2.0",
			expected: false,
		},
	}

	for _, tc := range cases {
		resp := isSupportedVersion(tc.input, mockBuildInfo)
		if resp != tc.expected {
			t.Errorf("expected %v, got %v", tc.expected, resp)
		}

		resp = isSupportedVersion(tc.input, &debug.BuildInfo{})
		if !resp {
			t.Errorf("expected %v, got %v", true, resp)
		}
	}
}

func TestParseSinceComment(t *testing.T) {
	cases := []struct {
		input              string
		expectedModuleName string
		expectedVersion    string
	}{
		{
			input:              "",
			expectedModuleName: "",
			expectedVersion:    "",
		},
		{
			input:              "not a since comment",
			expectedModuleName: "",
			expectedVersion:    "",
		},
		{
			input:              "//            Since: Cosmos SDK 0.50",
			expectedModuleName: "cosmos-sdk",
			expectedVersion:    "v0.50",
		},
		{
			input:              "// since: Cosmos SDK 0.50",
			expectedModuleName: "cosmos-sdk",
			expectedVersion:    "v0.50",
		},
		{
			input:              "// since: cosmos sdk 0.50",
			expectedModuleName: "cosmos-sdk",
			expectedVersion:    "v0.50",
		},
		{
			input:              "// since: Cosmos-SDK 0.50",
			expectedModuleName: "cosmos-sdk",
			expectedVersion:    "v0.50",
		},
		{
			input:              "// Since: cosmos-sdk v0.50",
			expectedModuleName: "cosmos-sdk",
			expectedVersion:    "v0.50",
		},
		{
			input:              "//since: cosmos-sdk v0.50.1",
			expectedModuleName: "cosmos-sdk",
			expectedVersion:    "v0.50.1",
		},
		{
			input:              "// since: cosmos-sdk 0.47.0-veronica",
			expectedModuleName: "cosmos-sdk",
			expectedVersion:    "v0.47.0-veronica",
		},
		{
			input:              "// Since: x/feegrant v0.1.0",
			expectedModuleName: "x/feegrant",
			expectedVersion:    "v0.1.0",
		},
		{
			input:              "// since: x/feegrant 0.1",
			expectedModuleName: "x/feegrant",
			expectedVersion:    "v0.1",
		},
	}

	for _, tc := range cases {
		moduleName, version := parseSinceComment(tc.input)
		if moduleName != tc.expectedModuleName {
			t.Errorf("expected module name %s, got %s", tc.expectedModuleName, moduleName)
		}
		if version != tc.expectedVersion {
			t.Errorf("expected version %s, got %s", tc.expectedVersion, version)
		}
	}
}
