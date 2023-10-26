package secp256k1

import (
	"math/big"
	"testing"

	secp "github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/stretchr/testify/require"
)

func Test_genPrivKey(t *testing.T) {
	empty := make([]byte, 32)
	oneB := big.NewInt(1).Bytes()
	onePadded := make([]byte, 32)
	copy(onePadded[32-len(oneB):32], oneB)
	t.Logf("one padded: %v, len=%v", onePadded, len(onePadded))

	validOne := append(empty, onePadded...)
	tests := []struct {
		name        string
		notSoRand   []byte
		shouldPanic bool
	}{
		{"valid because 0 < 1 < N", validOne, false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				require.Panics(t, func() {
					genPrivKey()
				})
				return
			}
			got := genPrivKey()
			fe := new(big.Int).SetBytes(got)
			require.True(t, fe.Cmp(secp.S256().N) < 0)
			require.True(t, fe.Sign() > 0)
		})
	}
}
