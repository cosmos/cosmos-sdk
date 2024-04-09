package ante

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections/colltest"
	"cosmossdk.io/core/appmodule"
	authcodec "cosmossdk.io/x/auth/codec"
	"cosmossdk.io/x/auth/keeper"
	authtypes "cosmossdk.io/x/auth/types"

	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestSigVerify_setPubKey(t *testing.T) {
	cdc := authcodec.NewBech32Codec("cosmos")

	maccPerms := map[string][]string{
		"fee_collector":          nil,
		"mint":                   {"minter"},
		"bonded_tokens_pool":     {"burner", "staking"},
		"not_bonded_tokens_pool": {"burner", "staking"},
		"multiPerm":              {"burner", "minter", "staking"},
		"random":                 {"random"},
	}
	store, _ := colltest.MockStore()
	authorityAddr, err := cdc.BytesToString(authtypes.NewModuleAddress("gov"))
	require.NoError(t, err)
	svd := SigVerificationDecorator{ak: keeper.NewAccountKeeper(
		appmodule.Environment{KVStoreService: store}, codectestutil.CodecOptions{}.NewCodec(),
		authtypes.ProtoBaseAccount, maccPerms, cdc, sdk.Bech32MainPrefix, authorityAddr,
	)}

	alicePk := secp256k1.GenPrivKey().PubKey()
	bobPk := secp256k1.GenPrivKey().PubKey()

	aliceAddr, err := cdc.BytesToString(alicePk.Address())
	require.NoError(t, err)

	ctx := sdk.NewContext(nil, false, nil)

	t.Run("on not sig verify tx - skip", func(t *testing.T) {
		acc := &authtypes.BaseAccount{}
		ctx = ctx.WithExecMode(sdk.ExecModeSimulate).WithIsSigverifyTx(false)
		err := svd.setPubKey(ctx, acc, nil)
		require.NoError(t, err)
	})

	t.Run("on sim, populate with sim key, if pubkey is nil", func(t *testing.T) {
		acc := &authtypes.BaseAccount{Address: aliceAddr}
		ctx = ctx.WithExecMode(sdk.ExecModeSimulate).WithIsSigverifyTx(true)
		err := svd.setPubKey(ctx, acc, nil)
		require.NoError(t, err)
		require.Equal(t, acc.PubKey.GetCachedValue(), simSecp256k1Pubkey)
	})

	t.Run("on sim, populate with real pub key, if pubkey is not nil", func(t *testing.T) {
		acc := &authtypes.BaseAccount{Address: aliceAddr}
		ctx = ctx.WithExecMode(sdk.ExecModeSimulate).WithIsSigverifyTx(true)
		err := svd.setPubKey(ctx, acc, alicePk)
		require.NoError(t, err)
		require.Equal(t, acc.PubKey.GetCachedValue(), alicePk)
	})

	t.Run("not on sim, populate the address", func(t *testing.T) {
		acc := &authtypes.BaseAccount{Address: aliceAddr}
		ctx = ctx.WithExecMode(sdk.ExecModeFinalize).WithIsSigverifyTx(true)
		err := svd.setPubKey(ctx, acc, alicePk)
		require.NoError(t, err)
		require.Equal(t, acc.PubKey.GetCachedValue(), alicePk)
	})

	t.Run("not on sim, fail on invalid pubkey.address", func(t *testing.T) {
		acc := &authtypes.BaseAccount{Address: aliceAddr}
		ctx = ctx.WithExecMode(sdk.ExecModeFinalize).WithIsSigverifyTx(true)
		err := svd.setPubKey(ctx, acc, bobPk)
		require.ErrorContains(t, err, "cannot be claimed")
	})
}
