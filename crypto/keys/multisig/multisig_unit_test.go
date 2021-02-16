package multisig

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/stretchr/testify/require"
)

func TestCheckKeysUnique(t *testing.T) {
	pk1 := secp256k1.GenPrivKey().PubKey()
	pk2 := secp256k1.GenPrivKey().PubKey()
	cases := []struct {
		name string
		pks  []cryptotypes.PubKey
		eq   bool
	}{
		{"same keys", []cryptotypes.PubKey{pk1, pk1}, false},
		{"same keys2", []cryptotypes.PubKey{pk1, pk1, pk1}, false},
		{"with duplicate", []cryptotypes.PubKey{pk1, pk2, pk1}, false},
		{"single key", []cryptotypes.PubKey{pk1}, true},
		{"diff keys", []cryptotypes.PubKey{pk1, pk2}, true},
	}
	for i := range cases {
		tc := cases[i]
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.eq, checkKeysUnique(tc.pks))
		})
	}
}
