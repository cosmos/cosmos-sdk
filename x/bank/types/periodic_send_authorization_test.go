package types_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

func TestPeriodicSendAuthorization(t *testing.T) {
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{
		Time: time.Now(),
	})
	now := ctx.BlockTime()
	inOne := now.Add(1 * time.Minute)
	oneHour := now.Add(1 * time.Hour)
	tenMinutes := time.Duration(10) * time.Minute
	totalLimit := sdk.NewCoins(sdk.NewInt64Coin("atom", 2000))
	periodLimit := sdk.NewCoins(sdk.NewInt64Coin("atom", 50))

	authorization := types.NewPeriodicSendAuthorization(
		totalLimit,
		&oneHour,
		tenMinutes,
		periodLimit,
		inOne,
	)

	t.Log("verify authorization returns valid method name")
	require.Equal(t, authorization.MsgTypeURL(), "/cosmos.bank.v1beta1.MsgSend")
	require.NoError(t, authorization.ValidateBasic())

	t.Log("verify period allowed")
	send := types.NewMsgSend(fromAddr, toAddr, sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(40))))
	require.NoError(t, authorization.ValidateBasic())
	resp, err := authorization.Accept(ctx, send)
	require.NoError(t, err)
	require.False(t, resp.Delete)
	authorization = resp.Updated.(*types.PeriodicSendAuthorization)
	left := sdk.NewCoins(sdk.NewInt64Coin("atom", 10))
	t.Log(authorization.PeriodCanSpend.AmountOf("atom"))
	t.Log(left.AmountOf("atom"))
	require.True(t, authorization.PeriodCanSpend.AmountOf("atom").Equal(left.AmountOf("atom")))

	t.Log("verify period limit exceeded")
	send = types.NewMsgSend(fromAddr, toAddr, sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(20))))
	resp, err = authorization.Accept(ctx, send)
	require.Error(t, err)
	require.False(t, resp.Delete)

	t.Log("verify period reset")
	send = types.NewMsgSend(fromAddr, toAddr, sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(20))))
	ctx = app.BaseApp.NewContext(false, tmproto.Header{}).WithBlockTime(now.Add(2 * time.Minute))
	resp, err = authorization.Accept(ctx, send)
	require.NoError(t, err)
	require.False(t, resp.Delete)

	t.Log("verify delete")
	send = types.NewMsgSend(fromAddr, toAddr, sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(20))))
	ctx = app.BaseApp.NewContext(false, tmproto.Header{}).WithBlockTime(now.Add(4 * time.Hour))
	resp, err = authorization.Accept(ctx, send)
	require.Error(t, err)
	require.True(t, resp.Delete)

}
