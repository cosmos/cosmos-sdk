package ante

import (
	"testing"

	cmtmldsa65 "github.com/cometbft/cometbft/crypto/mldsa65"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec/legacy"
	"github.com/cosmos/cosmos-sdk/crypto/keys/mldsa65"
	kmultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256r1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
)

// TestSimSigTxSize pins the simulated signature size for each key type,
// including the summed-subkey multisig sizing, which only exceeds the legacy
// TxSigLimit multiplier when TxSigLimit is small.
func TestSimSigTxSize(t *testing.T) {
	// stdSig is the encoded size of a placeholder signature of sigSize bytes.
	stdSig := func(pk cryptotypes.PubKey, sigSize int) storetypes.Gas {
		return storetypes.Gas(len(legacy.Cdc.MustMarshal(legacytx.StdSignature{ //nolint:staticcheck // SA1019: legacytx.StdSignature is deprecated
			Signature: make([]byte, sigSize),
			PubKey:    pk,
		})) + 6)
	}

	mlPks := make([]cryptotypes.PubKey, 3)
	for i := range mlPks {
		sk, err := mldsa65.GenPrivKey()
		require.NoError(t, err)
		mlPks[i] = sk.PubKey()
	}
	mlMultisig := kmultisig.NewLegacyAminoPubKey(2, mlPks)

	secpPks := make([]cryptotypes.PubKey, 3)
	for i := range secpPks {
		secpPks[i] = secp256k1.GenPrivKey().PubKey()
	}
	secpMultisig := kmultisig.NewLegacyAminoPubKey(2, secpPks)
	secpPk := secp256k1.GenPrivKey().PubKey()
	secpR1Sk, err := secp256r1.GenPrivKey()
	require.NoError(t, err)

	const secpSigSize = 64

	testCases := []struct {
		name       string
		pubkey     cryptotypes.PubKey
		txSigLimit uint64
		want       storetypes.Gas
	}{
		{"secp256k1", secpPk, 7, stdSig(secpPk, secpSigSize)},
		{"secp256r1", secpR1Sk.PubKey(), 7, stdSig(secpPk, secpSigSize)},
		{"ml-dsa-65", mlPks[0], 7, stdSig(mlPks[0], cmtmldsa65.SignatureSize)},
		{"secp256k1 multisig", secpMultisig, 7, 7 * stdSig(secpMultisig, secpSigSize)},
		// With the default limit, the legacy multiplier dominates.
		{"ml-dsa-65 multisig, txSigLimit 7", mlMultisig, 7, 7 * stdSig(mlMultisig, secpSigSize)},
		// With a limit of 1, the summed subkey signature sizes decide.
		{"ml-dsa-65 multisig, txSigLimit 1", mlMultisig, 1, stdSig(mlMultisig, 3*cmtmldsa65.SignatureSize)},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, simSigTxSize(tc.pubkey, tc.txSigLimit))
		})
	}
}
