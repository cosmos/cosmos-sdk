package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

func TestCacheWrapCodec(t *testing.T) {
	defaultCdc := codec.DefaultPublicKeyCodec{}
	wrappedCdc := types.CacheWrapCodec(defaultCdc)

	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey().(secp256k1.PubKeySecp256k1)

	pubKeyEnc, err := wrappedCdc.Encode(pubKey)
	require.NoError(t, err)
	require.Equal(t, pubKey, pubKeyEnc.GetCachedPubKey())

	pubKeyEnc2 := &types.PublicKey{
		Sum: &types.PublicKey_Secp256K1{Secp256K1: pubKey[:]},
	}
	pubKey2, err := wrappedCdc.Decode(pubKeyEnc2)
	require.NoError(t, err)
	require.Equal(t, pubKey, pubKey2)
	require.Equal(t, pubKey, pubKeyEnc2.GetCachedPubKey())
}
