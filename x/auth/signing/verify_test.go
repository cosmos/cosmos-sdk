package signing_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

func TestVerifySignature(t *testing.T) {
	priv, pubKey, addr := testdata.KeyTestPubAddr()
	priv1, pubKey1, addr1 := testdata.KeyTestPubAddr()

	const (
		memo    = "testmemo"
		chainId = "test-chain"
	)

	app, ctx := createTestApp(false)
	ctx = ctx.WithBlockHeight(1)

	cdc := codec.New()
	sdk.RegisterCodec(cdc)
	types.RegisterCodec(cdc)
	cdc.RegisterConcrete(testdata.TestMsg{}, "cosmos-sdk/Test", nil)

	acc1 := app.AccountKeeper.NewAccountWithAddress(ctx, addr)
	_ = app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	app.AccountKeeper.SetAccount(ctx, acc1)
	balances := sdk.NewCoins(sdk.NewInt64Coin("atom", 200))
	require.NoError(t, app.BankKeeper.SetBalances(ctx, addr, balances))
	acc, err := ante.GetSignerAcc(ctx, app.AccountKeeper, addr)
	require.NoError(t, app.BankKeeper.SetBalances(ctx, addr, balances))

	msgs := []sdk.Msg{testdata.NewTestMsg(addr)}
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
	require.NoError(t, err)

	handler := MakeTestHandlerMap()
	stdTx := types.NewStdTx(msgs, fee, []types.StdSignature{stdSig}, memo)
	err = signing.VerifySignature(pubKey, signerData, sigV2.Data, handler, stdTx)
	require.NoError(t, err)

	pkSet := []crypto.PubKey{pubKey, pubKey1}
	multisigKey := multisig.NewPubKeyMultisigThreshold(2, pkSet)
	multisignature := multisig.NewMultisig(2)
	msgs = []sdk.Msg{testdata.NewTestMsg(addr, addr1)}
	multiSignBytes := types.StdSignBytes(signerData.ChainID, signerData.AccountNumber, signerData.AccountSequence,
		fee, msgs, memo)

	sig1, err := priv.Sign(multiSignBytes)
	require.NoError(t, err)
	stdSig1 := types.StdSignature{PubKey: pubKey.Bytes(), Signature: sig1}
	sig1V2, err := types.StdSignatureToSignatureV2(cdc, stdSig1)
	require.NoError(t, err)

	sig2, err := priv1.Sign(multiSignBytes)
	require.NoError(t, err)
	stdSig2 := types.StdSignature{PubKey: pubKey.Bytes(), Signature: sig2}
	sig2V2, err := types.StdSignatureToSignatureV2(cdc, stdSig2)
	require.NoError(t, err)

	err = multisig.AddSignatureFromPubKey(multisignature, sig1V2.Data, pkSet[0], pkSet)
	require.NoError(t, err)
	err = multisig.AddSignatureFromPubKey(multisignature, sig2V2.Data, pkSet[1], pkSet)
	require.NoError(t, err)

	err = signing.VerifySignature(multisigKey, signerData, multisignature, handler, stdTx)
	require.NoError(t, err)
}

// returns context and app with params set on account keeper
func createTestApp(isCheckTx bool) (*simapp.SimApp, sdk.Context) {
	app := simapp.Setup(isCheckTx)
	ctx := app.BaseApp.NewContext(isCheckTx, abci.Header{})
	app.AccountKeeper.SetParams(ctx, types.DefaultParams())

	return app, ctx
}
