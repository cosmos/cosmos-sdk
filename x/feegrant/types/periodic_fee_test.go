package types

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPeriodicFeeValidAllow(t *testing.T) {
	atom := sdk.NewCoins(sdk.NewInt64Coin("atom", 555))
	smallAtom := sdk.NewCoins(sdk.NewInt64Coin("atom", 43))
	leftAtom := sdk.NewCoins(sdk.NewInt64Coin("atom", 512))
	oneAtom := sdk.NewCoins(sdk.NewInt64Coin("atom", 1))
	eth := sdk.NewCoins(sdk.NewInt64Coin("eth", 1))

	cases := map[string]struct {
		allow PeriodicFeeAllowance
		// all other checks are ignored if valid=false
		fee           sdk.Coins
		blockTime     time.Time
		blockHeight   int64
		valid         bool
		accept        bool
		remove        bool
		remains       sdk.Coins
		remainsPeriod sdk.Coins
		periodReset   ExpiresAt
	}{
		"empty": {
			allow: PeriodicFeeAllowance{},
			valid: false,
		},
		"only basic": {
			allow: PeriodicFeeAllowance{
				Basic: BasicFeeAllowance{
					SpendLimit: atom,
					Expiration: ExpiresAtHeight(100),
				},
			},
			valid: false,
		},
		"empty basic": {
			allow: PeriodicFeeAllowance{
				Period:           BlockDuration(50),
				PeriodSpendLimit: smallAtom,
			},
			valid: false,
		},
		"mismatched currencies": {
			allow: PeriodicFeeAllowance{
				Basic: BasicFeeAllowance{
					SpendLimit: atom,
					Expiration: ExpiresAtHeight(100),
				},
				Period:           BlockDuration(10),
				PeriodSpendLimit: eth,
			},
			valid: false,
		},
		"first time": {
			allow: PeriodicFeeAllowance{
				Basic: BasicFeeAllowance{
					SpendLimit: atom,
					Expiration: ExpiresAtHeight(100),
				},
				Period:           BlockDuration(10),
				PeriodSpendLimit: smallAtom,
			},
			valid:         true,
			fee:           smallAtom,
			blockHeight:   75,
			accept:        true,
			remove:        false,
			remainsPeriod: nil,
			remains:       leftAtom,
			periodReset:   ExpiresAtHeight(85),
		},
		"same period": {
			allow: PeriodicFeeAllowance{
				Basic: BasicFeeAllowance{
					SpendLimit: atom,
					Expiration: ExpiresAtHeight(100),
				},
				Period:           BlockDuration(10),
				PeriodReset:      ExpiresAtHeight(80),
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
			periodReset:   ExpiresAtHeight(80),
		},
		"step one period": {
			allow: PeriodicFeeAllowance{
				Basic: BasicFeeAllowance{
					SpendLimit: atom,
					Expiration: ExpiresAtHeight(100),
				},
				Period:           BlockDuration(10),
				PeriodReset:      ExpiresAtHeight(70),
				PeriodSpendLimit: leftAtom,
			},
			valid:         true,
			fee:           leftAtom,
			blockHeight:   75,
			accept:        true,
			remove:        false,
			remainsPeriod: nil,
			remains:       smallAtom,
			periodReset:   ExpiresAtHeight(80), // one step from last reset, not now
		},
		"step limited by global allowance": {
			allow: PeriodicFeeAllowance{
				Basic: BasicFeeAllowance{
					SpendLimit: smallAtom,
					Expiration: ExpiresAtHeight(100),
				},
				Period:           BlockDuration(10),
				PeriodReset:      ExpiresAtHeight(70),
				PeriodSpendLimit: atom,
			},
			valid:         true,
			fee:           oneAtom,
			blockHeight:   75,
			accept:        true,
			remove:        false,
			remainsPeriod: smallAtom.Sub(oneAtom),
			remains:       smallAtom.Sub(oneAtom),
			periodReset:   ExpiresAtHeight(80), // one step from last reset, not now
		},
		"expired": {
			allow: PeriodicFeeAllowance{
				Basic: BasicFeeAllowance{
					SpendLimit: atom,
					Expiration: ExpiresAtHeight(100),
				},
				Period:           BlockDuration(10),
				PeriodSpendLimit: smallAtom,
			},
			valid:       true,
			fee:         smallAtom,
			blockHeight: 101,
			accept:      false,
			remove:      true,
		},
		"over period limit": {
			allow: PeriodicFeeAllowance{
				Basic: BasicFeeAllowance{
					SpendLimit: atom,
					Expiration: ExpiresAtHeight(100),
				},
				Period:           BlockDuration(10),
				PeriodReset:      ExpiresAtHeight(80),
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
			err := tc.allow.ValidateBasic()
			if !tc.valid {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// now try to deduct
			remove, err := tc.allow.Accept(tc.fee, tc.blockTime, tc.blockHeight)
			if !tc.accept {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			require.Equal(t, tc.remove, remove)
			if !remove {
				assert.Equal(t, tc.remains, tc.allow.Basic.SpendLimit)
				assert.Equal(t, tc.remainsPeriod, tc.allow.PeriodCanSpend)
				assert.Equal(t, tc.periodReset, tc.allow.PeriodReset)
			}
		})
	}
}
