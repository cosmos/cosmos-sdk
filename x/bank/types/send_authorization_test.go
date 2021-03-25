package types_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

var (
	posCoins = sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(1000)))
	fromAddr = sdk.AccAddress("_____from _____")
	toAddr   = sdk.AccAddress("_______to________")
)

func TestSendAuthorization(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	authorization := types.NewSendAuthorization(posCoins)
	require.Equal(t, authorization.MethodName(), "/cosmos.bank.v1beta1.Msg/Send")
	require.NoError(t, authorization.ValidateBasic())
	send := types.NewMsgSend(fromAddr, toAddr, posCoins)
	srvMsg := sdk.ServiceMsg{
		MethodName: "/cosmos.bank.v1beta1.Msg/Send",
		Request:    send,
	}
	require.NoError(t, authorization.ValidateBasic())
	updated, del, err := authorization.Accept(ctx, srvMsg)
	require.NoError(t, err)
	require.NotNil(t, del)
	require.Nil(t, updated)

}
