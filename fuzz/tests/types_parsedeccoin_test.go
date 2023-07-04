//go:build gofuzz || go1.18

package tests

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/types"
)

func FuzzTypesParseDecCoin(f *testing.F) {
	f.Fuzz(func(t *testing.T, data []byte) {
		_, err := types.ParseDecCoin(string(data))
		require.NoError(t, err)
	})
}
