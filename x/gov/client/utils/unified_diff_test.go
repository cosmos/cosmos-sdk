package utils_test

import (
	"strings"
	"testing"

	"github.com/cosmos/cosmos-sdk/x/gov/client/utils"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/require"
)

func TestGenerateUnifiedDiff(t *testing.T) {
	tests := []struct {
		name     string
		src      string
		dst      string
		expected string
	}{
		{
			name:     "No changes",
			src:      "Line one\nLine two\nLine three",
			dst:      "Line one\nLine two\nLine three",
			expected: ``,
		},
		{
			name: "Line added",
			src:  "Line one\nLine two",
			dst:  "Line one\nLine two\nLine three",
			expected: `@@ -1,2 +1,3 @@
 Line one
 Line two
+Line three
`,
		},
		{
			name: "Line deleted",
			src:  "Line one\nLine two\nLine three",
			dst:  "Line one\nLine three",
			expected: `@@ -1,3 +1,2 @@
 Line one
-Line two
 Line three
`,
		},
		{
			name: "Line modified",
			src:  "Line one\nLine two\nLine three",
			dst:  "Line one\nLine two modified\nLine three",
			expected: `@@ -1,3 +1,3 @@
 Line one
-Line two
+Line two modified
 Line three
`,
		},
		{
			name: "Multiple changes",
			src:  "Line one\nLine two\nLine three",
			dst:  "Line zero\nLine one\nLine three\nLine four",
			expected: `@@ -1,3 +1,4 @@
+Line zero
 Line one
-Line two
 Line three
+Line four
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff, err := utils.GenerateUnifiedDiff(tt.src, tt.dst)
			require.NoError(t, err)

			diffContent := strings.TrimPrefix(diff, "--- src\n+++ dst\n")
			expectedContent := strings.TrimPrefix(tt.expected, "--- src\n+++ dst\n")

			require.Equal(t, expectedContent, diffContent)
		})
	}
}

func TestUnifiedDiffIntegration(t *testing.T) {
	src := "Line one\nLine two\nLine three"
	dst := "Line zero\nLine one\nLine three\nLine four"

	diffStr, err := utils.GenerateUnifiedDiff(src, dst)
	require.NoError(t, err)

	result, err := types.ApplyUnifiedDiff(src, diffStr)
	require.NoError(t, err)
	require.Equal(t, dst, result)
}
