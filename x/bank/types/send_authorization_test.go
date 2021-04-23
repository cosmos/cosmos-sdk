package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

var (
	coins1000 = sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(1000)))
	coins500  = sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(500)))
	fromAddr  = sdk.AccAddress("_____from _____")
	toAddr    = sdk.AccAddress("_______to________")
)

func TestSendAuthorization(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	authorization := types.NewSendAuthorization(coins1000)

	t.Log("verify authorization returns valid method name")
	require.Equal(t, authorization.MethodName(), "/cosmos.bank.v1beta1.Msg/Send")
	require.NoError(t, authorization.ValidateBasic())
	send := types.NewMsgSend(fromAddr, toAddr, coins1000)
	srvMsg := sdk.ServiceMsg{
		MethodName: "/cosmos.bank.v1beta1.Msg/Send",
		Request:    send,
	}
	require.NoError(t, authorization.ValidateBasic())

	t.Log("verify updated authorization returns nil")
	updated, del, err := authorization.Accept(ctx, srvMsg)
	require.NoError(t, err)
	require.True(t, del)
	require.Nil(t, updated)

	authorization = types.NewSendAuthorization(coins1000)
	require.Equal(t, authorization.MethodName(), "/cosmos.bank.v1beta1.Msg/Send")
	require.NoError(t, authorization.ValidateBasic())
	send = types.NewMsgSend(fromAddr, toAddr, coins500)
	srvMsg = sdk.ServiceMsg{
		MethodName: "/cosmos.bank.v1beta1.Msg/Send",
		Request:    send,
	}
	require.NoError(t, authorization.ValidateBasic())
	updated, del, err = authorization.Accept(ctx, srvMsg)

	t.Log("verify updated authorization returns remaining spent limit")
	require.NoError(t, err)
	require.False(t, del)
	require.NotNil(t, updated)
	sendAuth := types.NewSendAuthorization(coins500)
	require.Equal(t, sendAuth.String(), updated.String())

	t.Log("expect updated authorization nil after spending remaining amount")
	updated, del, err = updated.Accept(ctx, srvMsg)
	require.NoError(t, err)
	require.True(t, del)
	require.Nil(t, updated)
}
