package signing_test

import (
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/tests"
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
	priv, pubKey, addr := types.KeyTestPubAddr()

	const (
		from = "test_sign"
		backend = "test"
		memo = "testmemo"
		chainId = "test-chain"
	)

	app, ctx := createTestApp(false)
	ctx = ctx.WithBlockHeight(1)

	dir, clean := tests.NewTestCaseDir(t)
	t.Cleanup(clean)

	path := hd.CreateHDPath(118, 0, 0).String()
	kr, err := keyring.New(sdk.KeyringServiceName(), backend, dir, nil)
	require.NoError(t, err)

	_, _, err = kr.NewMnemonic(from, keyring.English, path, hd.Secp256k1)
	require.NoError(t, err)

	cdc := codec.New()
	sdk.RegisterCodec(cdc)
	types.RegisterCodec(cdc)
	cdc.RegisterConcrete(sdk.TestMsg{}, "cosmos-sdk/Test", nil)

	acc1 := app.AccountKeeper.NewAccountWithAddress(ctx, addr)
	app.AccountKeeper.SetAccount(ctx, acc1)
	balances := sdk.NewCoins(sdk.NewInt64Coin("atom", 200))
	require.NoError(t, app.BankKeeper.SetBalances(ctx, addr, balances))
	acc, err := ante.GetSignerAcc(ctx, app.AccountKeeper, addr)
	require.NoError(t, app.BankKeeper.SetBalances(ctx, addr, balances))

	msgs := []sdk.Msg{types.NewTestMsg(addr)}
	fee := types.NewStdFee(50000, sdk.Coins{sdk.NewInt64Coin("atom", 150)})
	signerData := signing.SignerData{
		ChainID:         chainId,
		AccountNumber:   acc.GetAccountNumber(),
		AccountSequence: acc.GetSequence(),
	}
	signBytes := types.StdSignBytes(signerData.ChainID, signerData.AccountNumber, signerData.AccountSequence,
		fee, msgs, memo)
	signature, err := priv.Sign(signBytes)
	require.NoError(t, err)

	stdSig := types.StdSignature{PubKey: pubKey.Bytes(), Signature: signature}
	sigV2, err := types.StdSignatureToSignatureV2(cdc, stdSig)
	handler := MakeTestHandlerMap()
	stdTx := types.NewStdTx(msgs, fee, []types.StdSignature{stdSig}, memo)
	err = signing.VerifySignature(pubKey, signerData, sigV2.Data, handler, stdTx)
	require.NoError(t, err)
}

// returns context and app with params set on account keeper
func createTestApp(isCheckTx bool) (*simapp.SimApp, sdk.Context) {
	app := simapp.Setup(isCheckTx)
	ctx := app.BaseApp.NewContext(isCheckTx, abci.Header{})
	app.AccountKeeper.SetParams(ctx, types.DefaultParams())

	return app, ctx
}
