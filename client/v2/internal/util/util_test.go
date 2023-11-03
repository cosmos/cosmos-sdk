package util

import (
	"testing"
)

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
			input:              "// since: Cosmos-SDK 0.50",
			expectedModuleName: "cosmos-sdk",
			expectedVersion:    "0.50",
		},
		{
			input:              "// Since: cosmos-sdk v0.50",
			expectedModuleName: "cosmos-sdk",
			expectedVersion:    "0.50",
		},
		{
			input:              "// since: cosmos-sdk v0.50.1",
			expectedModuleName: "cosmos-sdk",
			expectedVersion:    "0.50.1",
		},
		{
			input:              "// Since: x/feegrant v0.1.0",
			expectedModuleName: "x/feegrant",
			expectedVersion:    "0.1.0",
		},
		{
			input:              "// since: x/feegrant 0.1",
			expectedModuleName: "x/feegrant",
			expectedVersion:    "0.1",
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
