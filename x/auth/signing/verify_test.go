package signing_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

func TestVerifySignature(t *testing.T) {
	msg := []byte{1, 2, 3, 4}
	_, pubKey, _ := types.KeyTestPubAddr()
	addr := sdk.AccAddress(pubKey.Address())
	app, ctx := createTestApp(false)
	ctx = ctx.WithBlockHeight(1)

	cdc := codec.New()
	sdk.RegisterCodec(cdc)
	types.RegisterCodec(cdc)

	acc1 := app.AccountKeeper.NewAccountWithAddress(ctx, addr)
	app.AccountKeeper.SetAccount(ctx, acc1)
	balances := sdk.NewCoins(sdk.NewInt64Coin("atom", 200))
	require.NoError(t, app.BankKeeper.SetBalances(ctx, addr, balances))

	fee := types.NewStdFee(50000, sdk.Coins{sdk.NewInt64Coin("atom", 150)})
	sig := types.StdSignature{PubKey: pubKey.Bytes(), Signature: msg}
	stdTx := types.NewStdTx([]sdk.Msg{types.NewTestMsg(addr)}, fee, []types.StdSignature{sig}, "testsigs")

	acc, err := ante.GetSignerAcc(ctx, app.AccountKeeper, addr)
	require.Nil(t, err)
	genesis := ctx.BlockHeight() == 0
	chainID := ctx.ChainID()
	var accNum uint64
	if !genesis {
		accNum = acc.GetAccountNumber()
	}
	signerData := signing.SignerData{
		ChainID:         chainID,
		AccountNumber:   accNum,
		AccountSequence: acc.GetSequence(),
	}

	sigV2, err := types.StdSignatureToSignatureV2(cdc, sig)
	handler := MakeTestHandlerMap()

	err = signing.VerifySignature(pubKey, signerData, sigV2.Data, handler, stdTx)
	t.Log(err)
	require.NoError(t, err)
}

// returns context and app with params set on account keeper
func createTestApp(isCheckTx bool) (*simapp.SimApp, sdk.Context) {
	app := simapp.Setup(isCheckTx)
	ctx := app.BaseApp.NewContext(isCheckTx, abci.Header{})
	app.AccountKeeper.SetParams(ctx, types.DefaultParams())

	return app, ctx
}
