//go:build gofuzz || go1.18

package tests

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func FuzzTypesVerifyAddressFormat(f *testing.F) {
	f.Fuzz(func(t *testing.T, data []byte) {
		err := types.VerifyAddressFormat(data)
		require.NoError(t, err)

	})
}
