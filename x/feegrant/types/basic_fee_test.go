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

func TestBasicFeeValidAllow(t *testing.T) {
	app := simapp.Setup(false)

	eth := sdk.NewCoins(sdk.NewInt64Coin("eth", 10))
	atom := sdk.NewCoins(sdk.NewInt64Coin("atom", 555))
	smallAtom := sdk.NewCoins(sdk.NewInt64Coin("atom", 43))
	bigAtom := sdk.NewCoins(sdk.NewInt64Coin("atom", 1000))
	leftAtom := sdk.NewCoins(sdk.NewInt64Coin("atom", 512))

	cases := map[string]struct {
		allowance *types.BasicAllowance
		// all other checks are ignored if valid=false
		fee         sdk.Coins
		blockHeight int64
		accept      bool
		remove     bool
		remains     sdk.Coins
	}{
		"empty": {
			allowance:  &types.BasicAllowance{},
			accept: true,
		},
		"small fee without expire": {
			allowance: &types.BasicAllowance{
				SpendLimit: atom,
			},
			fee:     smallAtom,
			accept:  true,
			remove:  false,
			remains: leftAtom,
		},
		"all fee without expire": {
			allowance: &types.BasicAllowance{
				SpendLimit: smallAtom,
			},
			fee:    smallAtom,
			accept: true,
			remove: true,
		},
		"wrong fee": {
			allowance: &types.BasicAllowance{
				SpendLimit: smallAtom,
			},
			fee:    eth,
			accept: false,
		},
		"non-expired": {
			allowance: &types.BasicAllowance{
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
			allowance: &types.BasicAllowance{
				SpendLimit: atom,
				Expiration: types.ExpiresAtHeight(100),
			},
			fee:         smallAtom,
			blockHeight: 121,
			accept:      false,
			remove:      true,
		},
		"fee more than allowed": {
			allowance: &types.BasicAllowance{
				SpendLimit: atom,
				Expiration: types.ExpiresAtHeight(100),
			},
			fee:         bigAtom,
			blockHeight: 85,
			accept:      false,
		},
		"with out spend limit": {
			allowance: &types.BasicAllowance{
				Expiration: types.ExpiresAtHeight(100),
			},
			fee:         bigAtom,
			blockHeight: 85,
			accept:      true,
		},
		"expired no spend limit": {
			allowance: &types.BasicAllowance{
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

func TestBasicFeeAllowTime(t *testing.T) {
	app := simapp.Setup(false)

	eth := sdk.NewCoins(sdk.NewInt64Coin("eth", 10))
	atom := sdk.NewCoins(sdk.NewInt64Coin("atom", 555))
	smallAtom := sdk.NewCoins(sdk.NewInt64Coin("atom", 43))
	bigAtom := sdk.NewCoins(sdk.NewInt64Coin("atom", 1000))
	leftAtom := sdk.NewCoins(sdk.NewInt64Coin("atom", 512))

	now := time.Now()
	oneHour := now.Add(1 * time.Hour)

	cases := map[string]struct {
		allow *types.BasicAllowance
		// all other checks are ignored if valid=false
		fee       sdk.Coins
		blockTime time.Time
		valid     bool
		accept    bool
		remove    bool
		remains   sdk.Coins
	}{
		"empty": {
			allow:  &types.BasicAllowance{},
			valid:  true,
			accept: true,
		},
		"small fee without expire": {
			allow: &types.BasicAllowance{
				SpendLimit: atom,
			},
			valid:   true,
			fee:     smallAtom,
			accept:  true,
			remove:  false,
			remains: leftAtom,
		},
		"all fee without expire": {
			allow: &types.BasicAllowance{
				SpendLimit: smallAtom,
			},
			valid:  true,
			fee:    smallAtom,
			accept: true,
			remove: true,
		},
		"wrong fee": {
			allow: &types.BasicAllowance{
				SpendLimit: smallAtom,
			},
			valid:  true,
			fee:    eth,
			accept: false,
		},
		"non-expired": {
			allow: &types.BasicAllowance{
				SpendLimit: atom,
				Expiration: types.ExpiresAtTime(oneHour),
			},
			valid:     true,
			fee:       smallAtom,
			blockTime: now,
			accept:    true,
			remove:    false,
			remains:   leftAtom,
		},
		"expired": {
			allow: &types.BasicAllowance{
				SpendLimit: atom,
				Expiration: types.ExpiresAtTime(now),
			},
			valid:     true,
			fee:       smallAtom,
			blockTime: oneHour,
			accept:    false,
			remove:    true,
		},
		"fee more than allowed": {
			allow: &types.BasicAllowance{
				SpendLimit: atom,
				Expiration: types.ExpiresAtTime(oneHour),
			},
			valid:     true,
			fee:       bigAtom,
			blockTime: now,
			accept:    false,
		},
		"without spend limit": {
			allow: &types.BasicAllowance{
				Expiration: types.ExpiresAtTime(oneHour),
			},
			valid:     true,
			fee:       bigAtom,
			blockTime: now,
			accept:    true,
		},
		"expired no spend limit": {
			allow: &types.BasicAllowance{
				Expiration: types.ExpiresAtTime(now),
			},
			valid:     true,
			fee:       bigAtom,
			blockTime: oneHour,
			accept:    false,
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
				assert.Equal(t, tc.allow.SpendLimit, tc.remains)
			}
		})
	}
}
