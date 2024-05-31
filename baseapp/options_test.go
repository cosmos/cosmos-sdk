package baseapp

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

// TestSetAppVersion verifies that SetAppVersion does not panic if paramStore is nil.
func TestSetAppVersion(t *testing.T) {
	baseApp := NewBaseApp("test", nil, nil, nil)
	baseApp.paramStore = nil // explicitly set to nil

	require.NotPanics(t, func() {
		baseApp.SetAppVersion(types.Context{}, uint64(1))
	})
}
