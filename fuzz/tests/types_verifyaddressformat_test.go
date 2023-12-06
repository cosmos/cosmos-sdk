//go:build gofuzz || go1.18

package tests

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/types/bech32"
)

func FuzzTypesVerifyAddressBech32(f *testing.F) {
	f.Fuzz(func(t *testing.T, data []byte) {
		_ = bech32.VerifyAddressBech32(data)
	})
}
