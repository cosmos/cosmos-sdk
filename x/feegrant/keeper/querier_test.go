package keeper_test

import (
	abci "github.com/tendermint/tendermint/abci/types"

	codec "github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	"github.com/cosmos/cosmos-sdk/x/feegrant/types"
)

func (suite *KeeperTestSuite) TestQuery() {
	ctx := suite.ctx
	k := suite.dk

	cdc := codec.New()
	types.RegisterCodec(cdc)

	// some helpers
	grant1 := types.FeeAllowanceGrant{
		Granter: suite.addr,
		Grantee: suite.addr3,
		Allowance: &types.BasicFeeAllowance{
			SpendLimit: sdk.NewCoins(sdk.NewInt64Coin("atom", 555)),
			Expiration: types.ExpiresAtHeight(334455),
		},
	}
	grant2 := types.FeeAllowanceGrant{
		Granter: suite.addr2,
		Grantee: suite.addr3,
		Allowance: &types.BasicFeeAllowance{
			SpendLimit: sdk.NewCoins(sdk.NewInt64Coin("eth", 123)),
			Expiration: types.ExpiresAtHeight(334455),
		},
	}

	// let's set up some initial state here
	k.GrantFeeAllowance(ctx, grant1)
	k.GrantFeeAllowance(ctx, grant2)

	// now try some queries
	cases := map[string]struct {
		path  []string
		valid bool
		res   []types.FeeAllowanceGrant
	}{
		"bad path": {
			path: []string{"foo", "bar"},
		},
		"no data": {
			// addr in bech32
			path:  []string{"fees", "cosmos157ez5zlaq0scm9aycwphhqhmg3kws4qusmekll"},
			valid: true,
		},
		"two grants": {
			// addr3 in bech32
			path:  []string{"fees", "cosmos1qk93t4j0yyzgqgt6k5qf8deh8fq6smpn3ntu3x"},
			valid: true,
			res:   []types.FeeAllowanceGrant{grant1, grant2},
		},
	}

	querier := keeper.NewQuerier(k)
	for name, tc := range cases {
		tc := tc
		suite.Run(name, func() {
			bz, err := querier(ctx, tc.path, abci.RequestQuery{})
			if !tc.valid {
				suite.Error(err)
				return
			}
			suite.NoError(err)

			var grants []types.FeeAllowanceGrant
			serr := cdc.UnmarshalJSON(bz, &grants)
			suite.NoError(serr)

			suite.Equal(tc.res, grants)
		})
	}

}
