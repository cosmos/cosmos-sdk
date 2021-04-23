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

func TestPeriodicFeeValidAllow(t *testing.T) {
	app := simapp.Setup(false)
	atom := sdk.NewCoins(sdk.NewInt64Coin("atom", 555))
	smallAtom := sdk.NewCoins(sdk.NewInt64Coin("atom", 43))
	leftAtom := sdk.NewCoins(sdk.NewInt64Coin("atom", 512))
	oneAtom := sdk.NewCoins(sdk.NewInt64Coin("atom", 1))
	eth := sdk.NewCoins(sdk.NewInt64Coin("eth", 1))

	cases := map[string]struct {
		allowance types.PeriodicFeeAllowance
		// all other checks are ignored if valid=false
		fee           sdk.Coins
		blockHeight   int64
		valid         bool
		accept        bool
		remove        bool
		remains       sdk.Coins
		remainsPeriod sdk.Coins
		periodReset   types.ExpiresAt
	}{
		"empty": {
			allowance: types.PeriodicFeeAllowance{},
			valid: false,
		},
		"only basic": {
			allowance: types.PeriodicFeeAllowance{
				Basic: types.BasicFeeAllowance{
					SpendLimit: atom,
					Expiration: types.ExpiresAtHeight(100),
				},
			},
			valid: false,
		},
		"empty basic": {
			allowance: types.PeriodicFeeAllowance{
				Period:           types.BlockDuration(10),
				PeriodSpendLimit: smallAtom,
				PeriodReset:      types.ExpiresAtHeight(70),
			},
			blockHeight: 75,
			valid:       true,
			accept:      true,
			remove:      false,
			periodReset: types.ExpiresAtHeight(80),
		},
		"mismatched currencies": {
			allowance: types.PeriodicFeeAllowance{
				Basic: types.BasicFeeAllowance{
					SpendLimit: atom,
					Expiration: types.ExpiresAtHeight(100),
				},
				Period:           types.BlockDuration(10),
				PeriodSpendLimit: eth,
			},
			valid: false,
		},
		"first time": {
			allowance: types.PeriodicFeeAllowance{
				Basic: types.BasicFeeAllowance{
					SpendLimit: atom,
					Expiration: types.ExpiresAtHeight(100),
				},
				Period:           types.BlockDuration(10),
				PeriodSpendLimit: smallAtom,
			},
			valid:         true,
			fee:           smallAtom,
			blockHeight:   75,
			accept:        true,
			remove:        false,
			remainsPeriod: nil,
			remains:       leftAtom,
			periodReset:   types.ExpiresAtHeight(85),
		},
		"same period": {
			allowance: types.PeriodicFeeAllowance{
				Basic: types.BasicFeeAllowance{
					SpendLimit: atom,
					Expiration: types.ExpiresAtHeight(100),
				},
				Period:           types.BlockDuration(10),
				PeriodReset:      types.ExpiresAtHeight(80),
				PeriodSpendLimit: leftAtom,
				PeriodCanSpend:   smallAtom,
			},
			valid:         true,
			fee:           smallAtom,
			blockHeight:   75,
			accept:        true,
			remove:        false,
			remainsPeriod: nil,
			remains:       leftAtom,
			periodReset:   types.ExpiresAtHeight(80),
		},
		"step one period": {
			allowance: types.PeriodicFeeAllowance{
				Basic: types.BasicFeeAllowance{
					SpendLimit: atom,
					Expiration: types.ExpiresAtHeight(100),
				},
				Period:           types.BlockDuration(10),
				PeriodReset:      types.ExpiresAtHeight(70),
				PeriodSpendLimit: leftAtom,
			},
			valid:         true,
			fee:           leftAtom,
			blockHeight:   75,
			accept:        true,
			remove:        false,
			remainsPeriod: nil,
			remains:       smallAtom,
			periodReset:   types.ExpiresAtHeight(80), // one step from last reset, not now
		},
		"step limited by global allowance": {
			allowance: types.PeriodicFeeAllowance{
				Basic: types.BasicFeeAllowance{
					SpendLimit: smallAtom,
					Expiration: types.ExpiresAtHeight(100),
				},
				Period:           types.BlockDuration(10),
				PeriodReset:      types.ExpiresAtHeight(70),
				PeriodSpendLimit: atom,
			},
			valid:         true,
			fee:           oneAtom,
			blockHeight:   75,
			accept:        true,
			remove:        false,
			remainsPeriod: smallAtom.Sub(oneAtom),
			remains:       smallAtom.Sub(oneAtom),
			periodReset:   types.ExpiresAtHeight(80), // one step from last reset, not now
		},
		"expired": {
			allowance: types.PeriodicFeeAllowance{
				Basic: types.BasicFeeAllowance{
					SpendLimit: atom,
					Expiration: types.ExpiresAtHeight(100),
				},
				Period:           types.BlockDuration(10),
				PeriodSpendLimit: smallAtom,
			},
			valid:       true,
			fee:         smallAtom,
			blockHeight: 101,
			accept:      false,
			remove:      true,
		},
		"over period limit": {
			allowance: types.PeriodicFeeAllowance{
				Basic: types.BasicFeeAllowance{
					SpendLimit: atom,
					Expiration: types.ExpiresAtHeight(100),
				},
				Period:           types.BlockDuration(10),
				PeriodReset:      types.ExpiresAtHeight(80),
				PeriodSpendLimit: leftAtom,
				PeriodCanSpend:   smallAtom,
			},
			valid:       true,
			fee:         leftAtom,
			blockHeight: 70,
			accept:      false,
			remove:      true,
		},
	}

	for name, stc := range cases {
		tc := stc // to make scopelint happy
		t.Run(name, func(t *testing.T) {
			err := tc.allowance.ValidateBasic()
			if !tc.valid {
				require.Error(t, err)
				return
			}
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
				assert.Equal(t, tc.remains, tc.allowance.Basic.SpendLimit)
				assert.Equal(t, tc.remainsPeriod, tc.allowance.PeriodCanSpend)
				assert.Equal(t, tc.periodReset, tc.allowance.PeriodReset)
			}
		})
	}
}
