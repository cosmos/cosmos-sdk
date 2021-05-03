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

func TestPeriodicFeeValidAllow(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{
		Time: time.Now(),
	})

	atom := sdk.NewCoins(sdk.NewInt64Coin("atom", 555))
	smallAtom := sdk.NewCoins(sdk.NewInt64Coin("atom", 43))
	leftAtom := sdk.NewCoins(sdk.NewInt64Coin("atom", 512))
	oneAtom := sdk.NewCoins(sdk.NewInt64Coin("atom", 1))
	eth := sdk.NewCoins(sdk.NewInt64Coin("eth", 1))

	now := ctx.BlockTime()
	oneHour := now.Add(1 * time.Hour)
	twoHours := now.Add(2 * time.Hour)
	tenMinutes := time.Duration(10) * time.Minute

	cases := map[string]struct {
		allow types.PeriodicAllowance
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
			allow: types.PeriodicAllowance{},
			valid: false,
		},
		"only basic": {
			allow: types.PeriodicAllowance{
				Basic: types.BasicAllowance{
					SpendLimit: atom,
					Expiration: &oneHour,
				},
			},
			valid: false,
		},
		"empty basic": {
			allow: types.PeriodicAllowance{
				Period:           tenMinutes,
				PeriodSpendLimit: smallAtom,
				PeriodReset:      now.Add(30 * time.Minute),
			},
			blockTime:   now,
			valid:       true,
			accept:      true,
			remove:      false,
			periodReset: now.Add(30 * time.Minute),
		},
		"mismatched currencies": {
			allow: types.PeriodicAllowance{
				Basic: types.BasicAllowance{
					SpendLimit: atom,
					Expiration: &oneHour,
				},
				Period:           tenMinutes,
				PeriodSpendLimit: eth,
			},
			valid: false,
		},
		"same period": {
			allow: types.PeriodicAllowance{
				Basic: types.BasicAllowance{
					SpendLimit: atom,
					Expiration: &twoHours,
				},
				Period:           tenMinutes,
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
			periodReset:   now.Add(1 * time.Hour),
		},
		"step one period": {
			allow: types.PeriodicAllowance{
				Basic: types.BasicAllowance{
					SpendLimit: atom,
					Expiration: &twoHours,
				},
				Period:           tenMinutes,
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
			allow: types.PeriodicAllowance{
				Basic: types.BasicAllowance{
					SpendLimit: smallAtom,
					Expiration: &twoHours,
				},
				Period:           tenMinutes,
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
			allow: types.PeriodicAllowance{
				Basic: types.BasicAllowance{
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
			allow: types.PeriodicAllowance{
				Basic: types.BasicAllowance{
					SpendLimit: atom,
					Expiration: &now,
				},
				Period:           time.Hour,
				PeriodReset:      now.Add(1 * time.Hour),
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
			err := tc.allow.ValidateBasic()
			if !tc.valid {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			ctx := app.BaseApp.NewContext(false, tmproto.Header{}).WithBlockTime(tc.blockTime)
			// now try to deduct
			remove, err := tc.allow.Accept(ctx, tc.fee, []sdk.Msg{})
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
