package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	codec "github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/fee_grant/exported"
	"github.com/cosmos/cosmos-sdk/x/fee_grant/internal/keeper"
	"github.com/cosmos/cosmos-sdk/x/fee_grant/internal/types"
)

type KeeperTestSuite struct {
	suite.Suite

	cdc *codec.Codec
	ctx sdk.Context
	dk  keeper.Keeper

	addr  sdk.AccAddress
	addr2 sdk.AccAddress
	addr3 sdk.AccAddress
	addr4 sdk.AccAddress
}

func (suite *KeeperTestSuite) SetupTest() {
	db := dbm.NewMemDB()

	cdc := codec.New()
	types.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	suite.cdc = cdc

	delCapKey := sdk.NewKVStoreKey("delKey")

	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(delCapKey, sdk.StoreTypeIAVL, db)
	ms.LoadLatestVersion()

	suite.dk = keeper.NewKeeper(cdc, delCapKey)
	suite.ctx = sdk.NewContext(ms, abci.Header{ChainID: "test-chain-id", Time: time.Now().UTC(), Height: 1234}, false, log.NewNopLogger())

	suite.addr = mustAddr("cosmos157ez5zlaq0scm9aycwphhqhmg3kws4qusmekll")
	suite.addr2 = mustAddr("cosmos1rjxwm0rwyuldsg00qf5lt26wxzzppjzxs2efdw")
	suite.addr3 = mustAddr("cosmos1qk93t4j0yyzgqgt6k5qf8deh8fq6smpn3ntu3x")
	suite.addr4 = mustAddr("cosmos1p9qh4ldfd6n0qehujsal4k7g0e37kel90rc4ts")

}

func mustAddr(acc string) sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(acc)
	if err != nil {
		panic(err)
	}
	return addr
}

func (suite *KeeperTestSuite) TestKeeperCrud(t *testing.T) {
	ctx := suite.ctx
	k := suite.dk

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
	k.GrantFeeAllowance(ctx, types.FeeAllowanceGrant{
		Granter: suite.addr, Grantee: suite.addr2, Allowance: &basic,
	})
	k.GrantFeeAllowance(ctx, types.FeeAllowanceGrant{
		Granter: suite.addr, Grantee: suite.addr3, Allowance: &basic2,
	})
	k.GrantFeeAllowance(ctx, types.FeeAllowanceGrant{
		Granter: suite.addr2, Grantee: suite.addr3, Allowance: &basic,
	})
	k.GrantFeeAllowance(ctx, types.FeeAllowanceGrant{
		Granter: suite.addr2, Grantee: suite.addr4, Allowance: &basic,
	})
	k.GrantFeeAllowance(ctx, types.FeeAllowanceGrant{
		Granter: suite.addr4, Grantee: suite.addr, Allowance: &basic2,
	})

	// remove some, overwrite other
	k.RevokeFeeAllowance(ctx, suite.addr, suite.addr2)
	k.RevokeFeeAllowance(ctx, suite.addr, suite.addr3)
	k.GrantFeeAllowance(ctx, types.FeeAllowanceGrant{
		Granter: suite.addr, Grantee: suite.addr3, Allowance: &basic,
	})
	k.GrantFeeAllowance(ctx, types.FeeAllowanceGrant{
		Granter: suite.addr2, Grantee: suite.addr3, Allowance: &basic2,
	})

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
			granter: suite.addr,
			grantee: suite.addr2,
		},
		"addr revoked and added": {
			granter:   suite.addr,
			grantee:   suite.addr3,
			allowance: &basic,
		},
		"addr never there": {
			granter: suite.addr,
			grantee: suite.addr4,
		},
		"addr modified": {
			granter:   suite.addr2,
			grantee:   suite.addr3,
			allowance: &basic2,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			allow := k.GetFeeAllowance(ctx, tc.granter, tc.grantee)
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
			grantee: suite.addr2,
		},
		"addr has one": {
			grantee: suite.addr,
			grants:  []types.FeeAllowanceGrant{{Granter: suite.addr4, Grantee: suite.addr, Allowance: &basic2}},
		},
		"addr3 has two": {
			grantee: suite.addr3,
			grants: []types.FeeAllowanceGrant{
				{Granter: suite.addr, Grantee: suite.addr3, Allowance: &basic},
				{Granter: suite.addr2, Grantee: suite.addr3, Allowance: &basic2},
			},
		},
	}

	for name, tc := range allCases {
		t.Run(name, func(t *testing.T) {
			var grants []types.FeeAllowanceGrant
			err := k.IterateAllGranteeFeeAllowances(ctx, tc.grantee, func(grant types.FeeAllowanceGrant) bool {
				grants = append(grants, grant)
				return false
			})
			require.NoError(t, err)
			assert.Equal(t, tc.grants, grants)
		})
	}
}

func (suite *KeeperTestSuite) TestUseGrantedFee(t *testing.T) {
	ctx := suite.ctx
	k := suite.dk

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
			granter: suite.addr,
			grantee: suite.addr2,
			fee:     atom,
			allowed: true,
			final:   nil,
		},
		"expired and removed": {
			granter: suite.addr,
			grantee: suite.addr3,
			fee:     eth,
			allowed: false,
			final:   nil,
		},
		"too high": {
			granter: suite.addr,
			grantee: suite.addr2,
			fee:     hugeAtom,
			allowed: false,
			final:   &future,
		},
		"use a little": {
			granter: suite.addr,
			grantee: suite.addr2,
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
			k.GrantFeeAllowance(ctx, types.FeeAllowanceGrant{
				Granter: suite.addr, Grantee: suite.addr2, Allowance: &future,
			})
			k.GrantFeeAllowance(ctx, types.FeeAllowanceGrant{
				Granter: suite.addr, Grantee: suite.addr3, Allowance: &expired,
			})

			allowed := k.UseGrantedFees(ctx, tc.granter, tc.grantee, tc.fee)
			require.Equal(t, tc.allowed, allowed)

			loaded := k.GetFeeAllowance(ctx, tc.granter, tc.grantee)
			require.Equal(t, tc.final, loaded)
		})
	}
}
