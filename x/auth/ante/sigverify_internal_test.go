package ante

import (
	"testing"

	authcodec "cosmossdk.io/x/auth/codec"
	authtypes "cosmossdk.io/x/auth/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestSigVerify_setPubKey(t *testing.T) {
	svd := SigVerificationDecorator{}

	alicePk := secp256k1.GenPrivKey().PubKey()
	bobPk := secp256k1.GenPrivKey().PubKey()

	cdc := authcodec.NewBech32Codec("cosmos")

	aliceAddr, err := cdc.BytesToString(alicePk.Address())
	require.NoError(t, err)

	ctx := sdk.Context{}.WithIsSigverifyTx(true)

	t.Run("on not sig verify tx - skip", func(t *testing.T) {
		acc := &authtypes.BaseAccount{}
		ctx := sdk.Context{}.WithIsSigverifyTx(false)
		err := svd.setPubKey(ctx, false, acc, nil)
		require.NoError(t, err)
	})

	t.Run("on sim, populate with sim key, if pubkey is nil", func(t *testing.T) {
		acc := &authtypes.BaseAccount{Address: aliceAddr}
		err := svd.setPubKey(ctx, true, acc, nil)
		require.NoError(t, err)
		require.Equal(t, acc.PubKey.GetCachedValue(), simSecp256k1Pubkey)
	})

	t.Run("on sim, populate with real pub key, if pubkey is not nil", func(t *testing.T) {
		acc := &authtypes.BaseAccount{Address: aliceAddr}
		err := svd.setPubKey(ctx, true, acc, alicePk)
		require.NoError(t, err)
		require.Equal(t, acc.PubKey.GetCachedValue(), alicePk)
	})

	t.Run("not on sim, populate the address", func(t *testing.T) {
		acc := &authtypes.BaseAccount{Address: aliceAddr}
		err := svd.setPubKey(ctx, false, acc, alicePk)
		require.NoError(t, err)
		require.Equal(t, acc.PubKey.GetCachedValue(), alicePk)
	})

	t.Run("not on sim, fail on invalid pubkey.address", func(t *testing.T) {
		acc := &authtypes.BaseAccount{Address: aliceAddr}
		err := svd.setPubKey(ctx, false, acc, bobPk)
		require.ErrorContains(t, err, "cannot be claimed")
	})

}
