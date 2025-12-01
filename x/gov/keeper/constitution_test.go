package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestApplyConstitutionAmendment(t *testing.T) {
	govKeeper, _, _, _, _, _, ctx := setupGovKeeper(t)

	tests := []struct {
		name                string
		initialConstitution string
		amendment           string
		expectedResult      string
		expectError         bool
	}{
		{
			name:                "failed patch application",
			initialConstitution: "Hello World",
			amendment:           "Hi World",
			expectError:         true,
		},
		{
			name:                "successful patch application",
			initialConstitution: "Hello\nWorld",
			amendment:           "@@ -1,2 +1,2 @@\n-Hello\n+Hi\n World",
			expectError:         false,
			expectedResult:      "Hi\nWorld",
		},
		{
			name:                "successful patch application with multiple hunks",
			initialConstitution: "Line one\nLine two\nLine three\nLine four\nLine five\nLine six\nLine seven\nLine eight\nLine nine",
			amendment:           "--- src\n+++ dst\n@@ -1,2 +1,2 @@\n-Line one\n+Line one modified\n Line two\n@@ -8,2 +8,2 @@\n Line eight\n-Line nine\n+Line nine modified",
			expectError:         false,
			expectedResult:      "Line one modified\nLine two\nLine three\nLine four\nLine five\nLine six\nLine seven\nLine eight\nLine nine modified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			govKeeper.Constitution.Set(ctx, tt.initialConstitution)
			updatedConstitution, err := govKeeper.ApplyConstitutionAmendment(ctx, tt.amendment)
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedResult, updatedConstitution)
			}
		})
	}
}
