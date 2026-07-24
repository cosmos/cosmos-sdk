package feegrant_test

import (
	"testing"
	"time"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
)

func TestPeriodicFeeValidAllow(t *testing.T) {
	key := storetypes.NewKVStoreKey(feegrant.StoreKey)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))

	ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Time: time.Now()})

	atom := sdk.NewCoins(sdk.NewInt64Coin("atom", 555))
	smallAtom := sdk.NewCoins(sdk.NewInt64Coin("atom", 43))
	leftAtom := sdk.NewCoins(sdk.NewInt64Coin("atom", 512))
	oneAtom := sdk.NewCoins(sdk.NewInt64Coin("atom", 1))
	eth := sdk.NewCoins(sdk.NewInt64Coin("eth", 1))
	emptyCoins := sdk.Coins{}

	now := ctx.BlockTime()
	oneHour := now.Add(1 * time.Hour)
	twoHours := now.Add(2 * time.Hour)
	tenMinutes := time.Duration(10) * time.Minute

	cases := map[string]struct {
		allow         feegrant.PeriodicAllowance
		fee           sdk.Coins
		blockTime     time.Time
		valid         bool // all other checks are ignored if valid=false
		accept        bool
		remove        bool
		remains       sdk.Coins
		remainsPeriod sdk.Coins
		periodReset   time.Time
	}{
		"empty": {
			allow: feegrant.PeriodicAllowance{},
			valid: false,
		},
		"only basic": {
			allow: feegrant.PeriodicAllowance{
				Basic: feegrant.BasicAllowance{
					SpendLimit: atom,
					Expiration: &oneHour,
				},
			},
			valid: false,
		},
		"empty basic": {
			allow: feegrant.PeriodicAllowance{
				Period:           tenMinutes,
				PeriodSpendLimit: smallAtom,
				PeriodReset:      now.Add(30 * time.Minute),
			},
			blockTime:     now,
			valid:         true,
			accept:        true,
			remove:        false,
			remainsPeriod: emptyCoins,
			periodReset:   now.Add(30 * time.Minute),
		},
		"mismatched currencies": {
			allow: feegrant.PeriodicAllowance{
				Basic: feegrant.BasicAllowance{
					SpendLimit: atom,
					Expiration: &oneHour,
				},
				Period:           tenMinutes,
				PeriodSpendLimit: eth,
			},
			valid: false,
		},
		"zero period": {
			// A zero period would make tryResetPeriod refill the spend limit on
			// every Accept, bypassing the per-period cap, so it must be rejected.
			allow: feegrant.PeriodicAllowance{
				Period:           time.Duration(0),
				PeriodSpendLimit: smallAtom,
			},
			valid: false,
		},
		"same period": {
			allow: feegrant.PeriodicAllowance{
				Basic: feegrant.BasicAllowance{
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
			remainsPeriod: emptyCoins,
			remains:       leftAtom,
			periodReset:   now.Add(1 * time.Hour),
		},
		"step one period": {
			allow: feegrant.PeriodicAllowance{
				Basic: feegrant.BasicAllowance{
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
			remainsPeriod: emptyCoins,
			remains:       smallAtom,
			periodReset:   oneHour.Add(tenMinutes), // one step from last reset, not now
		},
		"step limited by global allowance": {
			allow: feegrant.PeriodicAllowance{
				Basic: feegrant.BasicAllowance{
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
			remainsPeriod: smallAtom.Sub(oneAtom...),
			remains:       smallAtom.Sub(oneAtom...),
			periodReset:   oneHour.Add(tenMinutes), // one step from last reset, not now
		},
		"period reset no spend limit": {
			allow: feegrant.PeriodicAllowance{
				Period:           tenMinutes,
				PeriodReset:      now,
				PeriodSpendLimit: atom,
			},
			valid:         true,
			fee:           atom,
			blockTime:     oneHour,
			accept:        true,
			remove:        false,
			remainsPeriod: emptyCoins,
			periodReset:   oneHour.Add(tenMinutes), // one step from last reset, not now
		},
		"expired": {
			allow: feegrant.PeriodicAllowance{
				Basic: feegrant.BasicAllowance{
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
			allow: feegrant.PeriodicAllowance{
				Basic: feegrant.BasicAllowance{
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

			ctx := testCtx.Ctx.WithBlockTime(tc.blockTime)
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

// TestPeriodicFeeZeroPeriodBypassesCap documents the behavior that motivates
// rejecting a non-positive Period in ValidateBasic. With Period == 0,
// tryResetPeriod refills PeriodCanSpend on every Accept (PeriodReset never
// advances past the block time), so the per-period spend cap can be exceeded
// arbitrarily within a single block. Accept is exercised directly here because
// such an allowance can no longer pass ValidateBasic after the fix.
func TestPeriodicFeeZeroPeriodBypassesCap(t *testing.T) {
	key := storetypes.NewKVStoreKey(feegrant.StoreKey)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Time: time.Now()})

	periodLimit := sdk.NewCoins(sdk.NewInt64Coin("atom", 10))
	fee := sdk.NewCoins(sdk.NewInt64Coin("atom", 10))

	allowance := feegrant.PeriodicAllowance{
		Period:           time.Duration(0),
		PeriodSpendLimit: periodLimit,
	}

	// A zero period is rejected by validation, which is the actual guard.
	require.Error(t, allowance.ValidateBasic())

	// Demonstrate why: if such an allowance were accepted, the per-period cap of
	// 10 could be spent repeatedly because each Accept resets PeriodCanSpend.
	for i := 0; i < 3; i++ {
		remove, err := allowance.Accept(ctx, fee, []sdk.Msg{})
		require.NoError(t, err, "spend %d should not be rejected when Period == 0", i)
		require.False(t, remove)
	}
}
