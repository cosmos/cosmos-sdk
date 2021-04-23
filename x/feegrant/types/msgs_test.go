package types_test

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant/types"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestMsgs(t *testing.T) {
	addr, _ := sdk.AccAddressFromBech32("cosmos1aeuqja06474dfrj7uqsvukm6rael982kk89mqr")
	addr2, _ := sdk.AccAddressFromBech32("cosmos1nph3cfzk6trsmfxkeu943nvach5qw4vwstnvkl")
	atom := sdk.NewCoins(sdk.NewInt64Coin("atom", 555))
	basic := &types.BasicFeeAllowance{
		SpendLimit: atom,
		Expiration: types.ExpiresAtTime(time.Now().Add(3 * time.Hour)),
	}
	cases := map[string]struct {
		grantee sdk.AccAddress
		granter sdk.AccAddress
		grant 	*types.BasicFeeAllowance
		valid 	bool
	}{
		"valid":{
			grantee: addr,
			granter: addr2,
			grant: basic,
			valid: true,
		},
		"grantee == granter":{
			grantee: addr,
			granter: addr,
			grant: basic,
			valid: false,
		},
	}

	for _,tc := range cases {
		msg, err := types.NewMsgGrantFeeAllowance(tc.grant, tc.granter, tc.grantee)
		require.NoError(t, err)
		msgRevoke := types.NewMsgRevokeFeeAllowance(tc.granter, tc.grantee)
		valid := msg.ValidateBasic()
		validRevoke := msgRevoke.ValidateBasic()
		if tc.valid {
			require.NoError(t, valid)
			require.NoError(t, validRevoke)

			addrSlice := msg.GetSigners()
			require.Equal(t, tc.granter.String(), addrSlice[0].String())

			allowance, err := msg.GetFeeAllowanceI()
			require.NoError(t, err)
			require.Equal(t, tc.grant, allowance)

			addrSlice = msgRevoke.GetSigners()
			require.Equal(t, tc.granter.String(), addrSlice[0].String())

			cdc := codec.NewProtoCodec(codectypes.NewInterfaceRegistry())
			err = msg.UnpackInterfaces(cdc)
			require.NoError(t, err)
		} else {
			require.Error(t, valid)
			require.Error(t, validRevoke)
		}
	}
}
