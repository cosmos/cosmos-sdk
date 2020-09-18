package std_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"

	kmultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/std"
)

func roundTripTest(t *testing.T, pubKey crypto.PubKey) {
	cdc := std.DefaultPublicKeyCodec{}

	pubKeyEnc, err := cdc.Encode(pubKey)
	require.NoError(t, err)
	pubKeyDec, err := cdc.Decode(pubKeyEnc)
	require.NoError(t, err)
	require.Equal(t, pubKey, pubKeyDec)
}

func TestDefaultPublicKeyCodec(t *testing.T) {
	roundTripTest(t, nil)

	roundTripTest(t, crypto.PubKey(nil))

	pubKeySecp256k1 := secp256k1.GenPrivKey().PubKey()
	roundTripTest(t, pubKeySecp256k1)

	pubKeyMultisig := kmultisig.NewLegacyAminoPubKey(2, []crypto.PubKey{
		pubKeySecp256k1, secp256k1.GenPrivKey().PubKey(), secp256k1.GenPrivKey().PubKey(),
	})
	roundTripTest(t, pubKeyMultisig)
}
