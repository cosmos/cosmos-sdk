//go:build gofuzz || go1.18

package tests

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/types"
)

func FuzzTypesParseTimeBytes(f *testing.F) {
	f.Fuzz(func(t *testing.T, bin []byte) {
		ti, err := types.ParseTimeBytes(bin)
		if err != nil {
			return
		}
		brt := types.FormatTimeBytes(ti)
		if !bytes.Equal(brt, bin) {
			panic(fmt.Sprintf("Roundtrip failure, got\n%s", brt))
		}
	})
}
