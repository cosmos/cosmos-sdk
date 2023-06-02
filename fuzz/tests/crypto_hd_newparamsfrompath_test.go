//go:build gofuzz || go1.18

package tests

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/stretchr/testify/require"
)

func FuzzCryptoHDNewParamsFromPath(f *testing.F) {
	f.Fuzz(func(t *testing.T, data []byte) {
		_, err := hd.NewParamsFromPath(string(data))
		require.NoError(t, err)

	})
}
