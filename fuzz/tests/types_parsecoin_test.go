//go:build gofuzz || go1.18

package tests

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/types"
)

func FuzzTypesParseCoin(f *testing.F) {
	f.Fuzz(func(t *testing.T, data []byte) {
		_, err := types.ParseCoinNormalized(string(data))
		if err != nil {
			panic(err)
		}
	})
}
