package types_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant/types"
)

func TestBasicFeeValidAllow(t *testing.T) {
	app := simapp.Setup(false)

	eth := sdk.NewCoins(sdk.NewInt64Coin("eth", 10))
	atom := sdk.NewCoins(sdk.NewInt64Coin("atom", 555))
	smallAtom := sdk.NewCoins(sdk.NewInt64Coin("atom", 43))
	bigAtom := sdk.NewCoins(sdk.NewInt64Coin("atom", 1000))
	leftAtom := sdk.NewCoins(sdk.NewInt64Coin("atom", 512))

	cases := map[string]struct {
		allowance *types.BasicFeeAllowance
		// all other checks are ignored if valid=false
		fee         sdk.Coins
		blockHeight int64
		accept      bool
		remove     bool
		remains     sdk.Coins
	}{
		"empty": {
			allowance:  &types.BasicFeeAllowance{},
			accept: true,
		},
		"small fee without expire": {
			allowance: &types.BasicFeeAllowance{
				SpendLimit: atom,
			},
			fee:     smallAtom,
			accept:  true,
			remove:  false,
			remains: leftAtom,
		},
		"all fee without expire": {
			allowance: &types.BasicFeeAllowance{
				SpendLimit: smallAtom,
			},
			fee:    smallAtom,
			accept: true,
			remove: true,
		},
		"wrong fee": {
			allowance: &types.BasicFeeAllowance{
				SpendLimit: smallAtom,
			},
			fee:    eth,
			accept: false,
		},
		"non-expired": {
			allowance: &types.BasicFeeAllowance{
				SpendLimit: atom,
				Expiration: types.ExpiresAtHeight(100),
			},
			fee:         smallAtom,
			blockHeight: 85,
			accept:      true,
			remove:      false,
			remains:     leftAtom,
		},
		"expired": {
			allowance: &types.BasicFeeAllowance{
				SpendLimit: atom,
				Expiration: types.ExpiresAtHeight(100),
			},
			fee:         smallAtom,
			blockHeight: 121,
			accept:      false,
			remove:      true,
		},
		"fee more than allowed": {
			allowance: &types.BasicFeeAllowance{
				SpendLimit: atom,
				Expiration: types.ExpiresAtHeight(100),
			},
			fee:         bigAtom,
			blockHeight: 85,
			accept:      false,
		},
		"with out spend limit": {
			allowance: &types.BasicFeeAllowance{
				Expiration: types.ExpiresAtHeight(100),
			},
			fee:         bigAtom,
			blockHeight: 85,
			accept:      true,
		},
		"expired no spend limit": {
			allowance: &types.BasicFeeAllowance{
				Expiration: types.ExpiresAtHeight(100),
			},
			fee:         bigAtom,
			blockHeight: 120,
			accept:      false,
		},
	}

	for name, stc := range cases {
		tc := stc // to make scopelint happy
		t.Run(name, func(t *testing.T) {
			err := tc.allowance.ValidateBasic()
			require.NoError(t, err)

			ctx := app.BaseApp.NewContext(false, tmproto.Header{}).WithBlockHeight(tc.blockHeight)

			// now try to deduct
			removed, err := tc.allowance.Accept(ctx, tc.fee, []sdk.Msg{})
			if !tc.accept {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			require.Equal(t, tc.remove, removed)
			if !removed {
				assert.Equal(t, tc.allowance.SpendLimit, tc.remains)
			}
		})
	}
}
