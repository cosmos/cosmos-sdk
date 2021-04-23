package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant/types"
)

func TestGrant(t *testing.T) {
	app := simapp.Setup(false)
	addr, err := sdk.AccAddressFromBech32("cosmos1qk93t4j0yyzgqgt6k5qf8deh8fq6smpn3ntu3x")
	require.NoError(t, err)
	addr2, err := sdk.AccAddressFromBech32("cosmos1p9qh4ldfd6n0qehujsal4k7g0e37kel90rc4ts")
	require.NoError(t, err)
	atom := sdk.NewCoins(sdk.NewInt64Coin("atom", 555))
	zeroAtoms := sdk.NewCoins(sdk.NewInt64Coin("atom", 0))

	goodGrant, err := types.NewFeeAllowanceGrant(addr2, addr, &types.BasicFeeAllowance{
		SpendLimit: atom,
		Expiration: types.ExpiresAtHeight(100),
	})
	require.NoError(t, err)

	noGranteeGrant, err := types.NewFeeAllowanceGrant(addr2, nil, &types.BasicFeeAllowance{
		SpendLimit: atom,
		Expiration: types.ExpiresAtHeight(100),
	})
	require.NoError(t, err)

	noGranterGrant, err := types.NewFeeAllowanceGrant(nil, addr, &types.BasicFeeAllowance{
		SpendLimit: atom,
		Expiration: types.ExpiresAtHeight(100),
	})
	require.NoError(t, err)

	selfGrant, err := types.NewFeeAllowanceGrant(addr2, addr2, &types.BasicFeeAllowance{
		SpendLimit: atom,
		Expiration: types.ExpiresAtHeight(100),
	})
	require.NoError(t, err)

	badAllowanceGrant, err := types.NewFeeAllowanceGrant(addr2, addr, &types.BasicFeeAllowance{
		SpendLimit: atom,
		Expiration: types.ExpiresAtHeight(-1),
	})
	require.NoError(t, err)

	zeroAllowance, err := types.NewFeeAllowanceGrant(addr, addr2, &types.BasicFeeAllowance{
		SpendLimit: zeroAtoms,
		Expiration: types.ExpiresAtTime(time.Now().Add(3 * time.Hour)),
	})
	require.NoError(t, err)

	cdc := app.AppCodec()
	// RegisterLegacyAminoCodec(cdc)

	cases := map[string]struct {
		grant types.FeeAllowanceGrant
		valid bool
	}{
		"good": {
			grant: goodGrant,
			valid: true,
		},
		"no grantee": {
			grant: noGranteeGrant,
			valid: false,
		},
		"no granter": {
			grant: noGranterGrant,
			valid: false,
		},
		"self-grant": {
			grant: selfGrant,
			valid: false,
		},
		"bad allowance": {
			grant: badAllowanceGrant,
			valid: false,
		},
		"zero allowance": {
			grant: zeroAllowance,
			valid: false,
		},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			err := tc.grant.ValidateBasic()
			if !tc.valid {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// if it is valid, let's try to serialize, deserialize, and make sure it matches
			bz, err := cdc.MarshalBinaryBare(&tc.grant)
			require.NoError(t, err)
			var loaded types.FeeAllowanceGrant
			err = cdc.UnmarshalBinaryBare(bz, &loaded)
			require.NoError(t, err)

			err = loaded.ValidateBasic()
			require.NoError(t, err)

			assert.Equal(t, tc.grant, loaded)
		})
	}
}
