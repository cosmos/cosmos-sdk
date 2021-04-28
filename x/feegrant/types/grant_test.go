package types_test

import (
	"testing"
	"time"

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
	cdc := app.AppCodec()

	cases := map[string]struct {
		granter sdk.AccAddress
		grantee sdk.AccAddress
		limit sdk.Coins
		expires types.ExpiresAt
		valid bool
	}{
		"good": {
			granter: addr2,
			grantee: addr,
			limit: atom,
			expires: types.ExpiresAtHeight(100),
			valid: true,
		},
		"no grantee": {
			granter: addr2,
			grantee: nil,
			limit: atom,
			expires: types.ExpiresAtHeight(100),
			valid: false,
		},
		"no granter": {
			granter: nil,
			grantee: addr,
			limit: atom,
			expires: types.ExpiresAtHeight(100),
			valid: false,
		},
		"self-grant": {
			granter: addr2,
			grantee: addr2,
			limit: atom,
			expires: types.ExpiresAtHeight(100),
			valid: false,
		},
		"bad height": {
			granter: addr2,
			grantee: addr,
			limit: atom,
			expires: types.ExpiresAtHeight(-100),
			valid: false,
		},
		"zero allowance": {
			granter: addr2,
			grantee: addr,
			limit: zeroAtoms,
			expires: types.ExpiresAtTime(time.Now().Add(3 * time.Hour)),
			valid: false,
		},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			grant, err := types.NewGrant(tc.granter, tc.grantee, &types.BasicAllowance{
				SpendLimit: tc.limit,
				Expiration: tc.expires,
			})
			require.NoError(t, err)
			err = grant.ValidateBasic()

			if !tc.valid {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// if it is valid, let's try to serialize, deserialize, and make sure it matches
			bz, err := cdc.MarshalBinaryBare(&grant)
			require.NoError(t, err)
			var loaded types.Grant
			err = cdc.UnmarshalBinaryBare(bz, &loaded)
			require.NoError(t, err)

			err = loaded.ValidateBasic()
			require.NoError(t, err)
			require.Equal(t, grant, loaded)
		})
	}
}
