package types_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/simapp"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz/types"
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
		msgs       []sdk.ServiceMsg
		expectPass bool
	}{
		{"nil grantee address", nil, []sdk.ServiceMsg{}, false},
		{"valid test", grantee, []sdk.ServiceMsg{}, true},
		{"valid test: msg type", grantee, []sdk.ServiceMsg{
			{
				MethodName: types.SendAuthorization{}.MethodName(),
				Request: &banktypes.MsgSend{
					Amount:      sdk.NewCoins(sdk.NewInt64Coin("steak", 2)),
					FromAddress: granter.String(),
					ToAddress:   grantee.String(),
				},
			},
		}, true},
	}
	for i, tc := range tests {
		msg := types.NewMsgExecAuthorized(tc.grantee, tc.msgs)
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
		msg := types.NewMsgRevokeAuthorization(tc.granter, tc.grantee, tc.msgType)
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
		authorization    types.Authorization
		expiration       time.Time
		expectErr        bool
		expectPass       bool
	}{
		{"nil granter address", nil, grantee, &types.SendAuthorization{SpendLimit: coinsPos}, time.Now(), false, false},
		{"nil grantee address", granter, nil, &types.SendAuthorization{SpendLimit: coinsPos}, time.Now(), false, false},
		{"nil granter and grantee address", nil, nil, &types.SendAuthorization{SpendLimit: coinsPos}, time.Now(), false, false},
		{"nil authorization", granter, grantee, nil, time.Now(), true, false},
		{"valid test case", granter, grantee, &types.SendAuthorization{SpendLimit: coinsPos}, time.Now().AddDate(0, 1, 0), false, true},
		{"past time", granter, grantee, &types.SendAuthorization{SpendLimit: coinsPos}, time.Now().AddDate(0, 0, -1), false, false},
	}
	for i, tc := range tests {
		msg, err := types.NewMsgGrantAuthorization(
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

func TestMsgGrantAuthorizationGetSignBytes(t *testing.T) {
	period := time.Now().AddDate(0, 1, 0)
	expected := fmt.Sprintf(
		`{"type":"cosmos-sdk/MsgGrantAuthorization","value":{"authorization":{"type":"cosmos-sdk/SendAuthorization","value":{"spend_limit":[{"amount":"100","denom":"steak"}]}},"expiration":"%s","grantee":"cosmos1ta047h6lta0kwunpde6x2e2lta047h6l22453t","granter":"cosmos1ta047h6lta0kwunpde6x2ujlta047h6l3ksxz2"}}`,
		period.UTC().Format(time.RFC3339Nano))
	msg, err := types.NewMsgGrantAuthorization(
		granter, grantee, &types.SendAuthorization{SpendLimit: coinsPos}, period,
	)
	require.NoError(t, err)
	res := msg.GetSignBytes()
	require.Equal(t, expected, string(res))
}

func TestMsgRevokeAuthorizationGetSignBytes(t *testing.T) {
	expected := `{"type":"cosmos-sdk/MsgRevokeAuthorization","value":{"authorization_msg_type":"/cosmos.bank.v1beta1.Msg/Send","grantee":"cosmos1ta047h6lta0kwunpde6x2e2lta047h6l22453t","granter":"cosmos1ta047h6lta0kwunpde6x2ujlta047h6l3ksxz2"}}`
	msg := types.NewMsgRevokeAuthorization(
		granter, grantee, types.SendAuthorization{}.MethodName(),
	)
	res := msg.GetSignBytes()
	require.Equal(t, expected, string(res))
}

func TestMsgExecAuthorizedGetSignBytes(t *testing.T) {
	expected := `{"type":"cosmos-sdk/MsgExecAuthorized","value":{"grantee":"cosmos1ta047h6lta0kwunpde6x2e2lta047h6l22453t","msgs":[{"amount":[{"amount":"2","denom":"steak"}],"from_address":"cosmos1ta047h6lta0kwunpde6x2ujlta047h6l3ksxz2","to_address":"cosmos1ta047h6lta0kwunpde6x2e2lta047h6l22453t"}]}}`
	msg := types.NewMsgExecAuthorized(
		grantee, []sdk.ServiceMsg{
			{
				MethodName: types.SendAuthorization{}.MethodName(),
				Request: &banktypes.MsgSend{
					Amount:      sdk.NewCoins(sdk.NewInt64Coin("steak", 2)),
					FromAddress: granter.String(),
					ToAddress:   grantee.String(),
				},
			},
		},
	)
	var app *simapp.SimApp
	app = simapp.Setup(false)
	msg.UnpackInterfaces(app.AppCodec())
	res := msg.GetSignBytes()
	require.Equal(t, expected, string(res))
}
