package types_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/fee_grant/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBasicFeeValidAllow(t *testing.T) {
	eth := sdk.NewCoins(sdk.NewInt64Coin("eth", 10))
	atom := sdk.NewCoins(sdk.NewInt64Coin("atom", 555))
	smallAtom := sdk.NewCoins(sdk.NewInt64Coin("atom", 43))
	leftAtom := sdk.NewCoins(sdk.NewInt64Coin("atom", 512))

	cases := map[string]struct {
		allow *types.BasicFeeAllowance
		// all other checks are ignored if valid=false
		fee         sdk.Coins
		blockTime   time.Time
		blockHeight int64
		valid       bool
		accept      bool
		remove      bool
		remains     sdk.Coins
	}{
		"empty": {
			allow: &types.BasicFeeAllowance{},
			valid: false,
		},
		"small fee": {
			allow: &types.BasicFeeAllowance{
				SpendLimit: atom,
			},
			valid:   true,
			fee:     smallAtom,
			accept:  true,
			remove:  false,
			remains: leftAtom,
		},
		"all fee": {
			allow: &types.BasicFeeAllowance{
				SpendLimit: smallAtom,
			},
			valid:  true,
			fee:    smallAtom,
			accept: true,
			remove: true,
		},
		"wrong fee": {
			allow: &types.BasicFeeAllowance{
				SpendLimit: smallAtom,
			},
			valid:  true,
			fee:    eth,
			accept: false,
		},
		"non-expired": {
			allow: &types.BasicFeeAllowance{
				SpendLimit: atom,
				Expiration: types.ExpiresAtHeight(100),
			},
			valid:       true,
			fee:         smallAtom,
			blockHeight: 85,
			accept:      true,
			remove:      false,
			remains:     leftAtom,
		},
		"expired": {
			allow: &types.BasicFeeAllowance{
				SpendLimit: atom,
				Expiration: types.ExpiresAtHeight(100),
			},
			valid:       true,
			fee:         smallAtom,
			blockHeight: 121,
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
				assert.Equal(t, tc.allow.SpendLimit, tc.remains)
			}
		})
	}
}
