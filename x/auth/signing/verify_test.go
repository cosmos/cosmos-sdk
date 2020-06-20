package signing_test

import (
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/tests"
	"github.com/spf13/viper"
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
	_, pubKey, _ := types.KeyTestPubAddr()
	addr := sdk.AccAddress(pubKey.Address())
	app, ctx := createTestApp(false)
	ctx = ctx.WithBlockHeight(1)

	dir, clean := tests.NewTestCaseDir(t)
	t.Cleanup(clean)

	viper.Set(flags.FlagKeyringBackend, "test")
	path := hd.CreateHDPath(118, 0, 0).String()
	kr, err := keyring.New(sdk.KeyringServiceName(), "test", dir, nil)
	require.NoError(t, err)

	var from = "test_sign"
	_, _, err = kr.NewMnemonic(from, keyring.English, path, hd.Secp256k1)
	require.NoError(t, err)

	viper.Set(flags.FlagFrom, from)
	viper.Set(flags.FlagChainID, "test-chain")
	viper.Set(flags.FlagHome, dir)

	cdc := codec.New()
	sdk.RegisterCodec(cdc)
	types.RegisterCodec(cdc)
	cdc.RegisterConcrete(sdk.TestMsg{}, "cosmos-sdk/Test", nil)

	acc1 := app.AccountKeeper.NewAccountWithAddress(ctx, addr)
	app.AccountKeeper.SetAccount(ctx, acc1)
	balances := sdk.NewCoins(sdk.NewInt64Coin("atom", 200))
	require.NoError(t, app.BankKeeper.SetBalances(ctx, addr, balances))

	acc, err := ante.GetSignerAcc(ctx, app.AccountKeeper, addr)
	msgs := []sdk.Msg{types.NewTestMsg(addr)}

	fee := types.NewStdFee(50000, sdk.Coins{sdk.NewInt64Coin("atom", 150)})
	txBuilder := types.NewTxBuilder(types.DefaultTxEncoder(cdc), acc.GetAccountNumber(), acc.GetSequence(),
		200000, 1.1, false, "test-chain", "testsigs", sdk.NewCoins(),
		sdk.DecCoins{sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, sdk.NewDecWithPrec(10000, sdk.Precision))},
	).WithKeybase(kr)

	signBytes, err := txBuilder.BuildSignMsg(msgs)
	require.Nil(t, err)

	sign, err := txBuilder.Sign(from, signBytes)
	require.Nil(t, err)

	stdSig := types.StdSignature{PubKey: pubKey.Bytes(), Signature: sign}
	stdTx := types.NewStdTx(msgs, fee, []types.StdSignature{}, "testsigs")

	chainID := ctx.ChainID()

	signerData := signing.SignerData{
		ChainID:         chainID,
		AccountNumber:   acc.GetAccountNumber(),
		AccountSequence: acc.GetSequence(),
	}

	sigV2, err := types.StdSignatureToSignatureV2(cdc, stdSig)
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
