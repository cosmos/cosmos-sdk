package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	codec "github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/delegation/exported"
	"github.com/cosmos/cosmos-sdk/x/delegation/internal/keeper"
	"github.com/cosmos/cosmos-sdk/x/delegation/internal/types"
)

type testInput struct {
	cdc *codec.Codec
	ctx sdk.Context
	dk  keeper.Keeper
}

func setupTestInput() testInput {
	db := dbm.NewMemDB()

	cdc := codec.New()
	types.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)

	delCapKey := sdk.NewKVStoreKey("delKey")

	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(delCapKey, sdk.StoreTypeIAVL, db)
	ms.LoadLatestVersion()

	dk := keeper.NewKeeper(cdc, delCapKey)

	ctx := sdk.NewContext(ms, abci.Header{ChainID: "test-chain-id", Time: time.Now().UTC(), Height: 1234}, false, log.NewNopLogger())
	return testInput{cdc: cdc, ctx: ctx, dk: dk}
}

var (
	// some valid cosmos keys....
	addr  = mustAddr("cosmos157ez5zlaq0scm9aycwphhqhmg3kws4qusmekll")
	addr2 = mustAddr("cosmos1rjxwm0rwyuldsg00qf5lt26wxzzppjzxs2efdw")
	addr3 = mustAddr("cosmos1qk93t4j0yyzgqgt6k5qf8deh8fq6smpn3ntu3x")
	addr4 = mustAddr("cosmos1p9qh4ldfd6n0qehujsal4k7g0e37kel90rc4ts")
)

func mustAddr(acc string) sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(acc)
	if err != nil {
		panic(err)
	}
	return addr
}

func TestKeeperCrud(t *testing.T) {
	input := setupTestInput()
	ctx := input.ctx
	k := input.dk

	// some helpers
	atom := sdk.NewCoins(sdk.NewInt64Coin("atom", 555))
	eth := sdk.NewCoins(sdk.NewInt64Coin("eth", 123))
	basic := types.BasicFeeAllowance{
		SpendLimit: atom,
		Expiration: types.ExpiresAtHeight(334455),
	}
	basic2 := types.BasicFeeAllowance{
		SpendLimit: eth,
		Expiration: types.ExpiresAtHeight(172436),
	}

	// let's set up some initial state here
	err := k.DelegateFeeAllowance(ctx, types.FeeAllowanceGrant{
		Granter: addr, Grantee: addr2, Allowance: &basic,
	})
	require.NoError(t, err)
	err = k.DelegateFeeAllowance(ctx, types.FeeAllowanceGrant{
		Granter: addr, Grantee: addr3, Allowance: &basic2,
	})
	require.NoError(t, err)
	err = k.DelegateFeeAllowance(ctx, types.FeeAllowanceGrant{
		Granter: addr2, Grantee: addr3, Allowance: &basic,
	})
	require.NoError(t, err)
	err = k.DelegateFeeAllowance(ctx, types.FeeAllowanceGrant{
		Granter: addr2, Grantee: addr4, Allowance: &basic,
	})
	require.NoError(t, err)
	err = k.DelegateFeeAllowance(ctx, types.FeeAllowanceGrant{
		Granter: addr4, Grantee: addr, Allowance: &basic2,
	})
	require.NoError(t, err)

	// remove some, overwrite other
	k.RevokeFeeAllowance(ctx, addr, addr2)
	k.RevokeFeeAllowance(ctx, addr, addr3)
	err = k.DelegateFeeAllowance(ctx, types.FeeAllowanceGrant{
		Granter: addr, Grantee: addr3, Allowance: &basic,
	})
	require.NoError(t, err)
	err = k.DelegateFeeAllowance(ctx, types.FeeAllowanceGrant{
		Granter: addr2, Grantee: addr3, Allowance: &basic2,
	})
	require.NoError(t, err)

	// end state:
	// addr -> addr3 (basic)
	// addr2 -> addr3 (basic2), addr4(basic)
	// addr4 -> addr (basic2)

	// then lots of queries
	cases := map[string]struct {
		grantee   sdk.AccAddress
		granter   sdk.AccAddress
		allowance exported.FeeAllowance
	}{
		"addr revoked": {
			granter: addr,
			grantee: addr2,
		},
		"addr revoked and added": {
			granter:   addr,
			grantee:   addr3,
			allowance: &basic,
		},
		"addr never there": {
			granter: addr,
			grantee: addr4,
		},
		"addr modified": {
			granter:   addr2,
			grantee:   addr3,
			allowance: &basic2,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			allow, err := k.GetFeeAllowance(ctx, tc.granter, tc.grantee)
			require.NoError(t, err)
			if tc.allowance == nil {
				require.Nil(t, allow)
				return
			}
			require.NotNil(t, allow)
			require.Equal(t, tc.allowance, allow)
		})
	}

	allCases := map[string]struct {
		grantee sdk.AccAddress
		grants  []types.FeeAllowanceGrant
	}{
		"addr2 has none": {
			grantee: addr2,
		},
		"addr has one": {
			grantee: addr,
			grants:  []types.FeeAllowanceGrant{{Granter: addr4, Grantee: addr, Allowance: &basic2}},
		},
		"addr3 has two": {
			grantee: addr3,
			grants: []types.FeeAllowanceGrant{
				{Granter: addr, Grantee: addr3, Allowance: &basic},
				{Granter: addr2, Grantee: addr3, Allowance: &basic2},
			},
		},
	}

	for name, tc := range allCases {
		t.Run(name, func(t *testing.T) {
			grants, err := k.GetAllFeeAllowances(ctx, tc.grantee)
			require.NoError(t, err)
			assert.Equal(t, tc.grants, grants)
		})
	}
}

func TestUseDelegatedFee(t *testing.T) {
	input := setupTestInput()
	ctx := input.ctx
	k := input.dk

	// some helpers
	atom := sdk.NewCoins(sdk.NewInt64Coin("atom", 555))
	eth := sdk.NewCoins(sdk.NewInt64Coin("eth", 123))
	future := types.BasicFeeAllowance{
		SpendLimit: atom,
		Expiration: types.ExpiresAtHeight(5678),
	}
	expired := types.BasicFeeAllowance{
		SpendLimit: eth,
		Expiration: types.ExpiresAtHeight(55),
	}

	// for testing limits of the contract
	hugeAtom := sdk.NewCoins(sdk.NewInt64Coin("atom", 9999))
	smallAtom := sdk.NewCoins(sdk.NewInt64Coin("atom", 1))
	futureAfterSmall := types.BasicFeeAllowance{
		SpendLimit: sdk.NewCoins(sdk.NewInt64Coin("atom", 554)),
		Expiration: types.ExpiresAtHeight(5678),
	}

	// then lots of queries
	cases := map[string]struct {
		grantee sdk.AccAddress
		granter sdk.AccAddress
		fee     sdk.Coins
		allowed bool
		final   exported.FeeAllowance
	}{
		"use entire pot": {
			granter: addr,
			grantee: addr2,
			fee:     atom,
			allowed: true,
			final:   nil,
		},
		"expired and removed": {
			granter: addr,
			grantee: addr3,
			fee:     eth,
			allowed: false,
			final:   nil,
		},
		"too high": {
			granter: addr,
			grantee: addr2,
			fee:     hugeAtom,
			allowed: false,
			final:   &future,
		},
		"use a little": {
			granter: addr,
			grantee: addr2,
			fee:     smallAtom,
			allowed: true,
			final:   &futureAfterSmall,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			// let's set up some initial state here
			// addr -> addr2 (future)
			// addr -> addr3 (expired)
			err := k.DelegateFeeAllowance(ctx, types.FeeAllowanceGrant{
				Granter: addr, Grantee: addr2, Allowance: &future,
			})
			require.NoError(t, err)
			err = k.DelegateFeeAllowance(ctx, types.FeeAllowanceGrant{
				Granter: addr, Grantee: addr3, Allowance: &expired,
			})
			require.NoError(t, err)

			allowed := k.UseDelegatedFees(ctx, tc.granter, tc.grantee, tc.fee)
			require.Equal(t, tc.allowed, allowed)

			loaded, err := k.GetFeeAllowance(ctx, tc.granter, tc.grantee)
			require.NoError(t, err)
			require.Equal(t, tc.final, loaded)
		})
	}
}
