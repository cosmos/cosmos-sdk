package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant/types"
)

func TestPeriodicFeeValidAllowTime(t *testing.T) {
	app := simapp.Setup(false)
	atom := sdk.NewCoins(sdk.NewInt64Coin("atom", 555))
	smallAtom := sdk.NewCoins(sdk.NewInt64Coin("atom", 43))
	leftAtom := sdk.NewCoins(sdk.NewInt64Coin("atom", 512))
	oneAtom := sdk.NewCoins(sdk.NewInt64Coin("atom", 1))
	eth := sdk.NewCoins(sdk.NewInt64Coin("eth", 1))
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	now := ctx.BlockTime()
	thirtyMinutes := now.Add(30 * time.Minute)
	oneHour := now.Add(1 * time.Hour)
	twoHour := now.Add(2 * time.Hour)

	cases := map[string]struct {
		allowance types.PeriodicFeeAllowance
		// all other checks are ignored if valid=false
		fee           sdk.Coins
		blockTime     time.Time
		valid         bool
		accept        bool
		remove        bool
		remains       sdk.Coins
		remainsPeriod sdk.Coins
		periodReset   time.Time
	}{
		"empty": {
			allowance: types.PeriodicFeeAllowance{},
			valid:     false,
		},
		"only basic": {
			allowance: types.PeriodicFeeAllowance{
				Basic: types.BasicFeeAllowance{
					SpendLimit: atom,
					Expiration: &oneHour,
				},
			},
			valid: false,
		},
		"empty basic": {
			allow: types.PeriodicFeeAllowance{
				Period:           time.Duration(10) * time.Minute,
				PeriodSpendLimit: smallAtom,
				PeriodReset:      now.Add(30 * time.Minute),
			},
			blockTime:   now,
			valid:       true,
			accept:      true,
			remove:      false,
			periodReset: thirtyMinutes,
		},
		"mismatched currencies": {
			allowance: types.PeriodicFeeAllowance{
				Basic: types.BasicFeeAllowance{
					SpendLimit: atom,
					Expiration: &oneHour,
				},
				Period:           10 * time.Minute,
				PeriodSpendLimit: eth,
			},
			valid: false,
		},
		"same period": {
			allowance: types.PeriodicFeeAllowance{
				Basic: types.BasicFeeAllowance{
					SpendLimit: atom,
					Expiration: &twoHour,
				},
				Period:           10 * time.Minute,
				PeriodReset:      now.Add(1 * time.Hour),
				PeriodSpendLimit: leftAtom,
				PeriodCanSpend:   smallAtom,
			},
			valid:         true,
			fee:           smallAtom,
			blockTime:     now,
			accept:        true,
			remove:        false,
			remainsPeriod: nil,
			remains:       leftAtom,
			periodReset:   oneHour,
		},
		"step one period": {
			allowance: types.PeriodicFeeAllowance{
				Basic: types.BasicFeeAllowance{
					SpendLimit: atom,
					Expiration: &twoHour,
				},
				Period:           10 * time.Minute,
				PeriodReset:      now,
				PeriodSpendLimit: leftAtom,
			},
			valid:         true,
			fee:           leftAtom,
			blockTime:     now.Add(1 * time.Hour),
			accept:        true,
			remove:        false,
			remainsPeriod: nil,
			remains:       smallAtom,
			periodReset:   oneHour.Add(10 * time.Minute), // one step from last reset, not now
		},
		"step limited by global allowance": {
			allowance: types.PeriodicFeeAllowance{
				Basic: types.BasicFeeAllowance{
					SpendLimit: smallAtom,
					Expiration: &twoHour,
				},
				Period:           10 * time.Minute,
				PeriodReset:      now,
				PeriodSpendLimit: atom,
			},
			valid:         true,
			fee:           oneAtom,
			blockTime:     oneHour,
			accept:        true,
			remove:        false,
			remainsPeriod: smallAtom.Sub(oneAtom),
			remains:       smallAtom.Sub(oneAtom),
			periodReset:   oneHour.Add(10 * time.Minute), // one step from last reset, not now
		},
		"expired": {
			allowance: types.PeriodicFeeAllowance{
				Basic: types.BasicFeeAllowance{
					SpendLimit: atom,
					Expiration: &now,
				},
				Period:           time.Hour,
				PeriodSpendLimit: smallAtom,
			},
			valid:     true,
			fee:       smallAtom,
			blockTime: oneHour,
			accept:    false,
			remove:    true,
		},
		"over period limit": {
			allowance: types.PeriodicFeeAllowance{
				Basic: types.BasicFeeAllowance{
					SpendLimit: atom,
					Expiration: &twoHour,
				},
				Period:           time.Hour,
				PeriodReset:      oneHour,
				PeriodSpendLimit: leftAtom,
				PeriodCanSpend:   smallAtom,
			},
			valid:     true,
			fee:       leftAtom,
			blockTime: now,
			accept:    false,
			remove:    true,
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

			ctx := app.BaseApp.NewContext(false, tmproto.Header{}).WithBlockTime(tc.blockTime)
			// now try to deduct
			removed, err := tc.allowance.Accept(ctx, tc.fee, []sdk.Msg{})
			if !tc.accept {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			require.Equal(t, tc.remove, remove)
			if !remove {
				assert.Equal(t, tc.remains, tc.allow.Basic.SpendLimit)
				assert.Equal(t, tc.remainsPeriod, tc.allow.PeriodCanSpend)
				assert.Equal(t, tc.periodReset.String(), tc.allow.PeriodReset.String())
			}
		})
	}
}
