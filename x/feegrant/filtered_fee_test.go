package feegrant_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/x/feegrant"
	ocproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestFilteredFeeValidAllow(t *testing.T) {
	app := simapp.Setup(t, false)

	ctx := app.BaseApp.NewContext(false, ocproto.Header{})
	badTime := ctx.BlockTime().AddDate(0, 0, -1)
	allowace := &feegrant.BasicAllowance{
		Expiration: &badTime,
	}
	require.Error(t, allowace.ValidateBasic())

	ctx = app.BaseApp.NewContext(false, ocproto.Header{
		Time: time.Now(),
	})
	eth := sdk.NewCoins(sdk.NewInt64Coin("eth", 10))
	atom := sdk.NewCoins(sdk.NewInt64Coin("atom", 555))
	smallAtom := sdk.NewCoins(sdk.NewInt64Coin("atom", 43))
	bigAtom := sdk.NewCoins(sdk.NewInt64Coin("atom", 1000))
	leftAtom := sdk.NewCoins(sdk.NewInt64Coin("atom", 512))
	now := ctx.BlockTime()
	oneHour := now.Add(1 * time.Hour)

	cases := map[string]struct {
		allowance *feegrant.BasicAllowance
		msgs      []string
		// all other checks are ignored if valid=false
		fee       sdk.Coins
		blockTime time.Time
		valid     bool
		accept    bool
		remove    bool
		remains   sdk.Coins
	}{
		"msg contained": {
			allowance: &feegrant.BasicAllowance{},
			msgs:      []string{"/cosmos.feegrant.v1beta1.MsgRevokeAllowance"},
			accept:    true,
		},
		"msg not contained": {
			allowance: &feegrant.BasicAllowance{},
			msgs:      []string{"/cosmos.feegrant.v1beta1.MsgGrantAllowance"},
			accept:    false,
		},
		"small fee without expire": {
			allowance: &feegrant.BasicAllowance{
				SpendLimit: atom,
			},
			msgs:    []string{"/cosmos.feegrant.v1beta1.MsgRevokeAllowance"},
			fee:     smallAtom,
			accept:  true,
			remove:  false,
			remains: leftAtom,
		},
		"all fee without expire": {
			allowance: &feegrant.BasicAllowance{
				SpendLimit: smallAtom,
			},
			msgs:   []string{"/cosmos.feegrant.v1beta1.MsgRevokeAllowance"},
			fee:    smallAtom,
			accept: true,
			remove: true,
		},
		"wrong fee": {
			allowance: &feegrant.BasicAllowance{
				SpendLimit: smallAtom,
			},
			msgs:   []string{"/cosmos.feegrant.v1beta1.MsgRevokeAllowance"},
			fee:    eth,
			accept: false,
		},
		"non-expired": {
			allowance: &feegrant.BasicAllowance{
				SpendLimit: atom,
				Expiration: &oneHour,
			},
			msgs:      []string{"/cosmos.feegrant.v1beta1.MsgRevokeAllowance"},
			valid:     true,
			fee:       smallAtom,
			blockTime: now,
			accept:    true,
			remove:    false,
			remains:   leftAtom,
		},
		"expired": {
			allowance: &feegrant.BasicAllowance{
				SpendLimit: atom,
				Expiration: &now,
			},
			msgs:      []string{"/cosmos.feegrant.v1beta1.MsgRevokeAllowance"},
			valid:     true,
			fee:       smallAtom,
			blockTime: oneHour,
			accept:    false,
			remove:    true,
		},
		"fee more than allowed": {
			allowance: &feegrant.BasicAllowance{
				SpendLimit: atom,
				Expiration: &oneHour,
			},
			msgs:      []string{"/cosmos.feegrant.v1beta1.MsgRevokeAllowance"},
			valid:     true,
			fee:       bigAtom,
			blockTime: now,
			accept:    false,
		},
		"with out spend limit": {
			allowance: &feegrant.BasicAllowance{
				Expiration: &oneHour,
			},
			msgs:      []string{"/cosmos.feegrant.v1beta1.MsgRevokeAllowance"},
			valid:     true,
			fee:       bigAtom,
			blockTime: now,
			accept:    true,
		},
		"expired no spend limit": {
			allowance: &feegrant.BasicAllowance{
				Expiration: &now,
			},
			msgs:      []string{"/cosmos.feegrant.v1beta1.MsgRevokeAllowance"},
			valid:     true,
			fee:       bigAtom,
			blockTime: oneHour,
			accept:    false,
		},
	}

	for name, stc := range cases {
		tc := stc // to make scopelint happy
		t.Run(name, func(t *testing.T) {
			err := tc.allowance.ValidateBasic()
			require.NoError(t, err)

			ctx := app.BaseApp.NewContext(false, ocproto.Header{}).WithBlockTime(tc.blockTime)

			// create grant
			createGrant := func() feegrant.Grant {
				var granter, grantee sdk.AccAddress
				allowance, err := feegrant.NewAllowedMsgAllowance(tc.allowance, tc.msgs)
				require.NoError(t, err)
				grant, err := feegrant.NewGrant(granter, grantee, allowance)
				require.NoError(t, err)
				return grant
			}
			grant := createGrant()

			// create some msg
			call := feegrant.MsgRevokeAllowance{
				Granter: "",
				Grantee: "",
			}

			// now try to deduct
			allowance, err := grant.GetGrant()
			require.NoError(t, err)
			removed, err := allowance.Accept(ctx, tc.fee, []sdk.Msg{&call})
			if !tc.accept {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			require.Equal(t, tc.remove, removed)
			if !removed {
				updatedGrant := func(granter, grantee sdk.AccAddress,
					allowance feegrant.FeeAllowanceI) feegrant.Grant {
					newGrant, err := feegrant.NewGrant(
						granter,
						grantee,
						allowance)
					require.NoError(t, err)

					cdc := simapp.MakeTestEncodingConfig().Codec
					bz, err := cdc.Marshal(&newGrant)
					require.NoError(t, err)

					var loaded feegrant.Grant
					err = cdc.Unmarshal(bz, &loaded)
					require.NoError(t, err)
					return loaded
				}
				newGrant := updatedGrant(sdk.AccAddress(grant.Granter),
					sdk.AccAddress(grant.Grantee), allowance)

				newAllowance, err := newGrant.GetGrant()
				require.NoError(t, err)
				feeAllowance, err := newAllowance.(*feegrant.AllowedMsgAllowance).GetAllowance()
				require.NoError(t, err)
				assert.Equal(t, tc.remains, feeAllowance.(*feegrant.BasicAllowance).SpendLimit)
			}
		})
	}
}
