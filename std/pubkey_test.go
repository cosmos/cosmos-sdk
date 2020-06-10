package std_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/std"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/tendermint/tendermint/crypto/sr25519"

	"github.com/cosmos/cosmos-sdk/crypto/types/multisig"
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
	pubKeySecp256k1 := secp256k1.GenPrivKey().PubKey()
	roundTripTest(t, pubKeySecp256k1)

	pubKeyEd25519 := ed25519.GenPrivKey().PubKey()
	roundTripTest(t, pubKeyEd25519)

	pubKeySr25519 := sr25519.GenPrivKey().PubKey()
	roundTripTest(t, pubKeySr25519)

	pubKeyMultisig := multisig.NewPubKeyMultisigThreshold(2, []crypto.PubKey{
		pubKeySecp256k1, pubKeyEd25519, pubKeySr25519,
	})
	roundTripTest(t, pubKeyMultisig)
}
