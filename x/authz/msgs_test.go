package authz_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

var (
	coinsPos = sdk.NewCoins(sdk.NewInt64Coin("steak", 100))
	granter  = sdk.AccAddress("_______granter______")
	grantee  = sdk.AccAddress("_______grantee______")
)

func TestMsgExecAuthorized(t *testing.T) {
	tests := []struct {
		title      string
		grantee    sdk.AccAddress
		msgs       []sdk.Msg
		expectPass bool
	}{
		{"nil grantee address", nil, []sdk.Msg{}, false},
		{"zero-messages test: should fail", grantee, []sdk.Msg{}, false},
		{"valid test: msg type", grantee, []sdk.Msg{
			&banktypes.MsgSend{
				Amount:      sdk.NewCoins(sdk.NewInt64Coin("steak", 2)),
				FromAddress: granter.String(),
				ToAddress:   grantee.String(),
			},
		}, true},
	}
	for i, tc := range tests {
		msg := authz.NewMsgExec(tc.grantee, tc.msgs)
		if tc.expectPass {
			require.NoError(t, msg.ValidateBasic(), "test: %v", i)
		} else {
			require.Error(t, msg.ValidateBasic(), "test: %v", i)
		}
	}
}
func TestMsgRevokeAuthorization(t *testing.T) {
	tests := []struct {
		title            string
		granter, grantee sdk.AccAddress
		msgType          string
		expectPass       bool
	}{
		{"nil Granter address", nil, grantee, "hello", false},
		{"nil Grantee address", granter, nil, "hello", false},
		{"nil Granter and Grantee address", nil, nil, "hello", false},
		{"valid test case", granter, grantee, "hello", true},
	}
	for i, tc := range tests {
		msg := authz.NewMsgRevoke(tc.granter, tc.grantee, tc.msgType)
		if tc.expectPass {
			require.NoError(t, msg.ValidateBasic(), "test: %v", i)
		} else {
			require.Error(t, msg.ValidateBasic(), "test: %v", i)
		}
	}
}

func TestMsgGrantAuthorization(t *testing.T) {
	tests := []struct {
		title            string
		granter, grantee sdk.AccAddress
		authorization    authz.Authorization
		expiration       time.Time
		expectErr        bool
		expectPass       bool
	}{
		{"nil granter address", nil, grantee, &banktypes.SendAuthorization{SpendLimit: coinsPos}, time.Now(), false, false},
		{"nil grantee address", granter, nil, &banktypes.SendAuthorization{SpendLimit: coinsPos}, time.Now(), false, false},
		{"nil granter and grantee address", nil, nil, &banktypes.SendAuthorization{SpendLimit: coinsPos}, time.Now(), false, false},
		{"nil authorization", granter, grantee, nil, time.Now(), true, false},
		{"valid test case", granter, grantee, &banktypes.SendAuthorization{SpendLimit: coinsPos}, time.Now().AddDate(0, 1, 0), false, true},
		{"past time", granter, grantee, &banktypes.SendAuthorization{SpendLimit: coinsPos}, time.Now().AddDate(0, 0, -1), false, true}, // TODO need 0.45
	}
	for i, tc := range tests {
		msg, err := authz.NewMsgGrant(
			tc.granter, tc.grantee, tc.authorization, tc.expiration,
		)
		if !tc.expectErr {
			require.NoError(t, err)
		} else {
			continue
		}
		if tc.expectPass {
			require.NoError(t, msg.ValidateBasic(), "test: %v", i)
		} else {
			require.Error(t, msg.ValidateBasic(), "test: %v", i)
		}
	}
}

func TestMsgGrantGetAuthorization(t *testing.T) {
	require := require.New(t)

	m := authz.MsgGrant{}
	require.Nil(m.GetAuthorization())

	g := authz.GenericAuthorization{Msg: "some_type"}
	var err error
	m.Grant.Authorization, err = cdctypes.NewAnyWithValue(&g)
	require.NoError(err)
	require.Equal(m.GetAuthorization(), &g)

	g = authz.GenericAuthorization{Msg: "some_type2"}
	m.SetAuthorization(&g)
	require.Equal(m.GetAuthorization(), &g)
}
