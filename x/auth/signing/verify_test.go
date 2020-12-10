package signing_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	kmultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
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

	cdc := codec.NewLegacyAmino()
	sdk.RegisterLegacyAminoCodec(cdc)
	types.RegisterLegacyAminoCodec(cdc)
	cdc.RegisterConcrete(testdata.TestMsg{}, "cosmos-sdk/Test", nil)

	acc1 := app.AccountKeeper.NewAccountWithAddress(ctx, addr)
	_ = app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	app.AccountKeeper.SetAccount(ctx, acc1)
	balances := sdk.NewCoins(sdk.NewInt64Coin("atom", 200))
	require.NoError(t, app.BankKeeper.SetBalances(ctx, addr, balances))
	acc, err := ante.GetSignerAcc(ctx, app.AccountKeeper, addr)
	require.NoError(t, app.BankKeeper.SetBalances(ctx, addr, balances))

	msgs := []sdk.Msg{testdata.NewTestMsg(addr)}
	fee := legacytx.NewStdFee(50000, sdk.Coins{sdk.NewInt64Coin("atom", 150)})
	signerData := signing.SignerData{
		ChainID:       chainId,
		AccountNumber: acc.GetAccountNumber(),
		Sequence:      acc.GetSequence(),
	}
	signBytes := legacytx.StdSignBytes(signerData.ChainID, signerData.AccountNumber, signerData.Sequence, 10, fee, msgs, memo)
	signature, err := priv.Sign(signBytes)
	require.NoError(t, err)

	stdSig := legacytx.StdSignature{PubKey: pubKey, Signature: signature}
	sigV2, err := legacytx.StdSignatureToSignatureV2(cdc, stdSig)
	require.NoError(t, err)

	handler := MakeTestHandlerMap()
	stdTx := legacytx.NewStdTx(msgs, fee, []legacytx.StdSignature{stdSig}, memo)
	stdTx.TimeoutHeight = 10
	err = signing.VerifySignature(pubKey, signerData, sigV2.Data, handler, stdTx)
	require.NoError(t, err)

	pkSet := []cryptotypes.PubKey{pubKey, pubKey1}
	multisigKey := kmultisig.NewLegacyAminoPubKey(2, pkSet)
	multisignature := multisig.NewMultisig(2)
	msgs = []sdk.Msg{testdata.NewTestMsg(addr, addr1)}
	multiSignBytes := legacytx.StdSignBytes(signerData.ChainID, signerData.AccountNumber, signerData.Sequence, 10, fee, msgs, memo)

	sig1, err := priv.Sign(multiSignBytes)
	require.NoError(t, err)
	stdSig1 := legacytx.StdSignature{PubKey: pubKey, Signature: sig1}
	sig1V2, err := legacytx.StdSignatureToSignatureV2(cdc, stdSig1)
	require.NoError(t, err)

	sig2, err := priv1.Sign(multiSignBytes)
	require.NoError(t, err)
	stdSig2 := legacytx.StdSignature{PubKey: pubKey, Signature: sig2}
	sig2V2, err := legacytx.StdSignatureToSignatureV2(cdc, stdSig2)
	require.NoError(t, err)

	err = multisig.AddSignatureFromPubKey(multisignature, sig1V2.Data, pkSet[0], pkSet)
	require.NoError(t, err)
	err = multisig.AddSignatureFromPubKey(multisignature, sig2V2.Data, pkSet[1], pkSet)
	require.NoError(t, err)

	stdTx = legacytx.NewStdTx(msgs, fee, []legacytx.StdSignature{stdSig1, stdSig2}, memo)
	stdTx.TimeoutHeight = 10

	err = signing.VerifySignature(multisigKey, signerData, multisignature, handler, stdTx)
	require.NoError(t, err)
}

// returns context and app with params set on account keeper
func createTestApp(isCheckTx bool) (*simapp.SimApp, sdk.Context) {
	app := simapp.Setup(isCheckTx)
	ctx := app.BaseApp.NewContext(isCheckTx, tmproto.Header{})
	app.AccountKeeper.SetParams(ctx, types.DefaultParams())

	return app, ctx
}
