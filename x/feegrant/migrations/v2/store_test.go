package v2_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/core/header"
	coretesting "cosmossdk.io/core/testing"
	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/x/feegrant"
	v2 "cosmossdk.io/x/feegrant/migrations/v2"
	"cosmossdk.io/x/feegrant/module"

	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
)

func TestMigration(t *testing.T) {
	encodingConfig := moduletestutil.MakeTestEncodingConfig(codectestutil.CodecOptions{}, module.AppModule{})
	cdc := encodingConfig.Codec
	ac := addresscodec.NewBech32Codec("cosmos")

	ctx := testutil.DefaultContext(v2.ModuleName)
	granter1 := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	grantee1 := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	granter2 := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	grantee2 := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	spendLimit := sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1000)))
	now := ctx.HeaderInfo().Time
	oneDay := now.AddDate(0, 0, 1)
	twoDays := now.AddDate(0, 0, 2)

	grants := []struct {
		granter    sdk.AccAddress
		grantee    sdk.AccAddress
		spendLimit sdk.Coins
		expiration *time.Time
	}{
		{
			granter:    granter1,
			grantee:    grantee1,
			spendLimit: spendLimit,
			expiration: &twoDays,
		},
		{
			granter:    granter2,
			grantee:    grantee2,
			spendLimit: spendLimit,
			expiration: &oneDay,
		},
		{
			granter:    granter1,
			grantee:    grantee2,
			spendLimit: spendLimit,
		},
		{
			granter:    granter2,
			grantee:    grantee1,
			expiration: &oneDay,
		},
	}

	store := coretesting.KVStoreService(ctx, v2.ModuleName).OpenKVStore(ctx)
	for _, grant := range grants {
		granterStr, err := ac.BytesToString(grant.granter)
		require.NoError(t, err)
		granteeStr, err := ac.BytesToString(grant.grantee)
		require.NoError(t, err)
		newGrant, err := feegrant.NewGrant(granterStr, granteeStr, &feegrant.BasicAllowance{
			SpendLimit: grant.spendLimit,
			Expiration: grant.expiration,
		})
		require.NoError(t, err)

		bz, err := cdc.Marshal(&newGrant)
		require.NoError(t, err)

		err = store.Set(v2.FeeAllowanceKey(grant.granter, grant.grantee), bz)
		require.NoError(t, err)
	}

	ctx = ctx.WithHeaderInfo(header.Info{Time: now.Add(30 * time.Hour)})
	require.NoError(t, v2.MigrateStore(ctx, runtime.NewEnvironment(coretesting.KVStoreService(ctx, v2.ModuleName), coretesting.NewNopLogger()), cdc))

	s1, err := store.Get(v2.FeeAllowanceKey(granter1, grantee1))
	require.NoError(t, err)
	require.NotNil(t, s1)
	s2, err := store.Get(v2.FeeAllowanceKey(granter2, grantee2))
	require.NoError(t, err)
	require.NotNil(t, s2)
	s3, err := store.Get(v2.FeeAllowanceKey(granter1, grantee2))
	require.NoError(t, err)
	require.NotNil(t, s3)
	s4, err := store.Get(v2.FeeAllowanceKey(granter2, grantee1))
	require.NoError(t, err)
	require.NotNil(t, s4)
}
