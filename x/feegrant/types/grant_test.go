package types_test

import (
	"testing"

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

	goodGrant, err := types.NewGrant(addr2, addr, &types.BasicAllowance{
		SpendLimit: atom,
		Expiration: types.ExpiresAtHeight(100),
	})
	require.NoError(t, err)

	noGranteeGrant, err := types.NewGrant(addr2, nil, &types.BasicAllowance{
		SpendLimit: atom,
		Expiration: types.ExpiresAtHeight(100),
	})
	require.NoError(t, err)

	noGranterGrant, err := types.NewGrant(nil, addr, &types.BasicAllowance{
		SpendLimit: atom,
		Expiration: types.ExpiresAtHeight(100),
	})
	require.NoError(t, err)

	selfGrant, err := types.NewGrant(addr2, addr2, &types.BasicAllowance{
		SpendLimit: atom,
		Expiration: types.ExpiresAtHeight(100),
	})
	require.NoError(t, err)

	badAllowanceGrant, err := types.NewGrant(addr2, addr, &types.BasicAllowance{
		SpendLimit: atom,
		Expiration: types.ExpiresAtHeight(-1),
	})
	require.NoError(t, err)

	cdc := app.AppCodec()
	// RegisterLegacyAminoCodec(cdc)

	cases := map[string]struct {
		grant types.Grant
		valid bool
	}{
		"good": {
			grant: goodGrant,
			valid: true,
		},
		"no grantee": {
			grant: noGranteeGrant,
		},
		"no granter": {
			grant: noGranterGrant,
		},
		"self-grant": {
			grant: selfGrant,
		},
		"bad allowance": {
			grant: badAllowanceGrant,
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
			var loaded types.Grant
			err = cdc.UnmarshalBinaryBare(bz, &loaded)
			require.NoError(t, err)

			err = loaded.ValidateBasic()
			require.NoError(t, err)

			assert.Equal(t, tc.grant, loaded)
		})
	}
}
