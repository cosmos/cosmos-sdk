package util

import (
	"runtime/debug"
	"testing"

	"google.golang.org/protobuf/reflect/protoreflect"

	"cosmossdk.io/client/v2/internal/testpb"
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
		t.Run(tc.input, func(t *testing.T) {
			resp := isSupportedVersion(tc.input, mockBuildInfo)
			if resp != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, resp)
			}
		})
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
			expectedModuleName: "feegrant",
			expectedVersion:    "v0.1.0",
		},
		{
			input:              "// since: x/feegrant 0.1",
			expectedModuleName: "feegrant",
			expectedVersion:    "v0.1",
		},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			moduleName, version := parseSinceComment(tc.input)
			if moduleName != tc.expectedModuleName {
				t.Errorf("expected module name %s, got %s", tc.expectedModuleName, moduleName)
			}
			if version != tc.expectedVersion {
				t.Errorf("expected version %s, got %s", tc.expectedVersion, version)
			}
		})
	}
}

func TestDescriptorDocs(t *testing.T) {
	t.Skip() // TODO(@julienrbrt): Unskip when https://github.com/cosmos/cosmos-proto/pull/131 is finalized.

	msg1 := &testpb.MsgRequest{}
	descriptor1 := msg1.ProtoReflect().Descriptor()

	msg2 := testpb.MsgResponse{}
	descriptor2 := msg2.ProtoReflect().Descriptor()

	cases := []struct {
		name     string
		input    protoreflect.Descriptor
		expected string
	}{
		{
			name:     "Test with leading comments",
			input:    descriptor1,
			expected: "MsgRequest is a sample request message",
		},
		{
			name:     "Test with no leading comments",
			input:    descriptor2,
			expected: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			output := DescriptorDocs(tc.input)
			if output != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, output)
			}
		})
	}
}
