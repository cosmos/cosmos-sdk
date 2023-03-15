package ante_test

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	kmultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256r1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	txmodule "github.com/cosmos/cosmos-sdk/x/auth/tx/config"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func TestSetPubKey(t *testing.T) {
	suite := SetupTestSuite(t, true)
	suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()

	// keys and addresses
	priv1, pub1, addr1 := testdata.KeyTestPubAddr()
	priv2, pub2, addr2 := testdata.KeyTestPubAddr()
	priv3, pub3, addr3 := testdata.KeyTestPubAddrSecp256R1(t)

	addrs := []sdk.AccAddress{addr1, addr2, addr3}
	pubs := []cryptotypes.PubKey{pub1, pub2, pub3}

	msgs := make([]sdk.Msg, len(addrs))
	// set accounts and create msg for each address
	for i, addr := range addrs {
		acc := suite.accountKeeper.NewAccountWithAddress(suite.ctx, addr)
		require.NoError(t, acc.SetAccountNumber(uint64(i)))
		suite.accountKeeper.SetAccount(suite.ctx, acc)
		msgs[i] = testdata.NewTestMsg(addr)
	}
	require.NoError(t, suite.txBuilder.SetMsgs(msgs...))
	suite.txBuilder.SetFeeAmount(testdata.NewTestFeeAmount())
	suite.txBuilder.SetGasLimit(testdata.NewTestGasLimit())

	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1, priv2, priv3}, []uint64{0, 1, 2}, []uint64{0, 0, 0}
	tx, err := suite.CreateTestTx(suite.ctx, privs, accNums, accSeqs, suite.ctx.ChainID(), signing.SignMode_SIGN_MODE_DIRECT)
	require.NoError(t, err)

	spkd := ante.NewSetPubKeyDecorator(suite.accountKeeper)
	antehandler := sdk.ChainAnteDecorators(spkd)

	ctx, err := antehandler(suite.ctx, tx, false)
	require.NoError(t, err)

	// Require that all accounts have pubkey set after Decorator runs
	for i, addr := range addrs {
		pk, err := suite.accountKeeper.GetPubKey(ctx, addr)
		require.NoError(t, err, "Error on retrieving pubkey from account")
		require.True(t, pubs[i].Equals(pk),
			"Wrong Pubkey retrieved from AccountKeeper, idx=%d\nexpected=%s\n     got=%s", i, pubs[i], pk)
	}
}

func TestConsumeSignatureVerificationGas(t *testing.T) {
	suite := SetupTestSuite(t, true)
	params := types.DefaultParams()
	msg := []byte{1, 2, 3, 4}

	p := types.DefaultParams()
	skR1, _ := secp256r1.GenPrivKey()
	pkSet1, sigSet1 := generatePubKeysAndSignatures(5, msg, false)
	multisigKey1 := kmultisig.NewLegacyAminoPubKey(2, pkSet1)
	multisignature1 := multisig.NewMultisig(len(pkSet1))
	expectedCost1 := expectedGasCostByKeys(pkSet1)
	for i := 0; i < len(pkSet1); i++ {
		stdSig := legacytx.StdSignature{PubKey: pkSet1[i], Signature: sigSet1[i]} //nolint:staticcheck // SA1019: legacytx.StdSignature is deprecated
		sigV2, err := legacytx.StdSignatureToSignatureV2(suite.clientCtx.LegacyAmino, stdSig)
		require.NoError(t, err)
		err = multisig.AddSignatureV2(multisignature1, sigV2, pkSet1)
		require.NoError(t, err)
	}

	type args struct {
		meter  storetypes.GasMeter
		sig    signing.SignatureData
		pubkey cryptotypes.PubKey
		params types.Params
	}
	tests := []struct {
		name        string
		args        args
		gasConsumed uint64
		shouldErr   bool
	}{
		{"PubKeyEd25519", args{storetypes.NewInfiniteGasMeter(), nil, ed25519.GenPrivKey().PubKey(), params}, p.SigVerifyCostED25519, true},
		{"PubKeySecp256k1", args{storetypes.NewInfiniteGasMeter(), nil, secp256k1.GenPrivKey().PubKey(), params}, p.SigVerifyCostSecp256k1, false},
		{"PubKeySecp256r1", args{storetypes.NewInfiniteGasMeter(), nil, skR1.PubKey(), params}, p.SigVerifyCostSecp256r1(), false},
		{"Multisig", args{storetypes.NewInfiniteGasMeter(), multisignature1, multisigKey1, params}, expectedCost1, false},
		{"unknown key", args{storetypes.NewInfiniteGasMeter(), nil, nil, params}, 0, true},
	}
	for _, tt := range tests {
		sigV2 := signing.SignatureV2{
			PubKey:   tt.args.pubkey,
			Data:     tt.args.sig,
			Sequence: 0, // Arbitrary account sequence
		}
		err := ante.DefaultSigVerificationGasConsumer(tt.args.meter, sigV2, tt.args.params)

		if tt.shouldErr {
			require.NotNil(t, err)
		} else {
			require.Nil(t, err)
			require.Equal(t, tt.gasConsumed, tt.args.meter.GasConsumed(), fmt.Sprintf("%d != %d", tt.gasConsumed, tt.args.meter.GasConsumed()))
		}
	}
}

func TestSigVerification(t *testing.T) {
	suite := SetupTestSuite(t, true)
	suite.txBankKeeper.EXPECT().DenomMetadata(suite.ctx, gomock.Any()).Return(&banktypes.QueryDenomMetadataResponse{}, nil).AnyTimes()

	enabledSignModes := []signing.SignMode{signing.SignMode_SIGN_MODE_DIRECT, signing.SignMode_SIGN_MODE_TEXTUAL, signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON}
	// Since TEXTUAL is not enabled by default, we create a custom TxConfig
	// here which includes it.
	suite.clientCtx.TxConfig = authtx.NewTxConfigWithTextual(
		codec.NewProtoCodec(suite.encCfg.InterfaceRegistry),
		enabledSignModes,
		txmodule.NewTextualWithGRPCConn(suite.clientCtx),
	)
	suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()

	// make block height non-zero to ensure account numbers part of signBytes
	suite.ctx = suite.ctx.WithBlockHeight(1)

	// keys and addresses
	priv1, _, addr1 := testdata.KeyTestPubAddr()
	priv2, _, addr2 := testdata.KeyTestPubAddr()
	priv3, _, addr3 := testdata.KeyTestPubAddr()

	addrs := []sdk.AccAddress{addr1, addr2, addr3}

	msgs := make([]sdk.Msg, len(addrs))
	// set accounts and create msg for each address
	for i, addr := range addrs {
		acc := suite.accountKeeper.NewAccountWithAddress(suite.ctx, addr)
		require.NoError(t, acc.SetAccountNumber(uint64(i)))
		suite.accountKeeper.SetAccount(suite.ctx, acc)
		msgs[i] = testdata.NewTestMsg(addr)
	}

	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()

	spkd := ante.NewSetPubKeyDecorator(suite.accountKeeper)
	anteTxConfig := authtx.NewTxConfigWithTextual(
		codec.NewProtoCodec(suite.encCfg.InterfaceRegistry),
		enabledSignModes,
		txmodule.NewTextualWithBankKeeper(suite.txBankKeeper),
	)
	svd := ante.NewSigVerificationDecorator(suite.accountKeeper, anteTxConfig.SignModeHandler())
	antehandler := sdk.ChainAnteDecorators(spkd, svd)

	type testCase struct {
		name        string
		privs       []cryptotypes.PrivKey
		accNums     []uint64
		accSeqs     []uint64
		invalidSigs bool // used for testing sigverify on RecheckTx
		recheck     bool
		shouldErr   bool
	}
	validSigs := false
	testCases := []testCase{
		{"no signers", []cryptotypes.PrivKey{}, []uint64{}, []uint64{}, validSigs, false, true},
		{"not enough signers", []cryptotypes.PrivKey{priv1, priv2}, []uint64{0, 1}, []uint64{0, 0}, validSigs, false, true},
		{"wrong order signers", []cryptotypes.PrivKey{priv3, priv2, priv1}, []uint64{2, 1, 0}, []uint64{0, 0, 0}, validSigs, false, true},
		{"wrong accnums", []cryptotypes.PrivKey{priv1, priv2, priv3}, []uint64{7, 8, 9}, []uint64{0, 0, 0}, validSigs, false, true},
		{"wrong sequences", []cryptotypes.PrivKey{priv1, priv2, priv3}, []uint64{0, 1, 2}, []uint64{3, 4, 5}, validSigs, false, true},
		{"valid tx", []cryptotypes.PrivKey{priv1, priv2, priv3}, []uint64{0, 1, 2}, []uint64{0, 0, 0}, validSigs, false, false},
		{"no err on recheck", []cryptotypes.PrivKey{priv1, priv2, priv3}, []uint64{0, 0, 0}, []uint64{0, 0, 0}, !validSigs, true, false},
	}

	for i, tc := range testCases {
		for _, signMode := range enabledSignModes {
			t.Run(fmt.Sprintf("%s with %s", tc.name, signMode), func(t *testing.T) {
				suite.ctx = suite.ctx.WithIsReCheckTx(tc.recheck)
				suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder() // Create new txBuilder for each test

				require.NoError(t, suite.txBuilder.SetMsgs(msgs...))
				suite.txBuilder.SetFeeAmount(feeAmount)
				suite.txBuilder.SetGasLimit(gasLimit)

				tx, err := suite.CreateTestTx(suite.ctx, tc.privs, tc.accNums, tc.accSeqs, suite.ctx.ChainID(), signMode)
				require.NoError(t, err)
				if tc.invalidSigs {
					txSigs, _ := tx.GetSignaturesV2()
					badSig, _ := tc.privs[0].Sign([]byte("unrelated message"))
					txSigs[0] = signing.SignatureV2{
						PubKey: tc.privs[0].PubKey(),
						Data: &signing.SingleSignatureData{
							SignMode:  suite.clientCtx.TxConfig.SignModeHandler().DefaultMode(),
							Signature: badSig,
						},
						Sequence: tc.accSeqs[0],
					}
					suite.txBuilder.SetSignatures(txSigs...)
					tx = suite.txBuilder.GetTx()
				}

				_, err = antehandler(suite.ctx, tx, false)
				if tc.shouldErr {
					require.NotNil(t, err, "TestCase %d: %s did not error as expected", i, tc.name)
				} else {
					require.Nil(t, err, "TestCase %d: %s errored unexpectedly. Err: %v", i, tc.name, err)
				}
			})
		}
	}
}

func TestSigIntegration(t *testing.T) {
	// generate private keys
	privs := []cryptotypes.PrivKey{
		secp256k1.GenPrivKey(),
		secp256k1.GenPrivKey(),
		secp256k1.GenPrivKey(),
	}

	params := types.DefaultParams()
	initialSigCost := params.SigVerifyCostSecp256k1
	initialCost, err := runSigDecorators(t, params, false, privs...)
	require.Nil(t, err)

	params.SigVerifyCostSecp256k1 *= 2
	doubleCost, err := runSigDecorators(t, params, false, privs...)
	require.Nil(t, err)

	require.Equal(t, initialSigCost*uint64(len(privs)), doubleCost-initialCost)
}

func runSigDecorators(t *testing.T, params types.Params, _ bool, privs ...cryptotypes.PrivKey) (storetypes.Gas, error) {
	suite := SetupTestSuite(t, true)
	suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()

	// Make block-height non-zero to include accNum in SignBytes
	suite.ctx = suite.ctx.WithBlockHeight(1)
	err := suite.accountKeeper.SetParams(suite.ctx, params)
	require.NoError(t, err)

	msgs := make([]sdk.Msg, len(privs))
	accNums := make([]uint64, len(privs))
	accSeqs := make([]uint64, len(privs))
	// set accounts and create msg for each address
	for i, priv := range privs {
		addr := sdk.AccAddress(priv.PubKey().Address())
		acc := suite.accountKeeper.NewAccountWithAddress(suite.ctx, addr)
		require.NoError(t, acc.SetAccountNumber(uint64(i)))
		suite.accountKeeper.SetAccount(suite.ctx, acc)
		msgs[i] = testdata.NewTestMsg(addr)
		accNums[i] = uint64(i)
		accSeqs[i] = uint64(0)
	}
	require.NoError(t, suite.txBuilder.SetMsgs(msgs...))

	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()
	suite.txBuilder.SetFeeAmount(feeAmount)
	suite.txBuilder.SetGasLimit(gasLimit)

	tx, err := suite.CreateTestTx(suite.ctx, privs, accNums, accSeqs, suite.ctx.ChainID(), signing.SignMode_SIGN_MODE_DIRECT)
	require.NoError(t, err)

	spkd := ante.NewSetPubKeyDecorator(suite.accountKeeper)
	svgc := ante.NewSigGasConsumeDecorator(suite.accountKeeper, ante.DefaultSigVerificationGasConsumer)
	svd := ante.NewSigVerificationDecorator(suite.accountKeeper, suite.clientCtx.TxConfig.SignModeHandler())
	antehandler := sdk.ChainAnteDecorators(spkd, svgc, svd)

	// Determine gas consumption of antehandler with default params
	before := suite.ctx.GasMeter().GasConsumed()
	ctx, err := antehandler(suite.ctx, tx, false)
	after := ctx.GasMeter().GasConsumed()

	return after - before, err
}

func TestIncrementSequenceDecorator(t *testing.T) {
	suite := SetupTestSuite(t, true)
	suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()

	priv, _, addr := testdata.KeyTestPubAddr()
	acc := suite.accountKeeper.NewAccountWithAddress(suite.ctx, addr)
	require.NoError(t, acc.SetAccountNumber(uint64(50)))
	suite.accountKeeper.SetAccount(suite.ctx, acc)

	msgs := []sdk.Msg{testdata.NewTestMsg(addr)}
	require.NoError(t, suite.txBuilder.SetMsgs(msgs...))
	privs := []cryptotypes.PrivKey{priv}
	accNums := []uint64{suite.accountKeeper.GetAccount(suite.ctx, addr).GetAccountNumber()}
	accSeqs := []uint64{suite.accountKeeper.GetAccount(suite.ctx, addr).GetSequence()}
	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()
	suite.txBuilder.SetFeeAmount(feeAmount)
	suite.txBuilder.SetGasLimit(gasLimit)

	tx, err := suite.CreateTestTx(suite.ctx, privs, accNums, accSeqs, suite.ctx.ChainID(), signing.SignMode_SIGN_MODE_DIRECT)
	require.NoError(t, err)

	isd := ante.NewIncrementSequenceDecorator(suite.accountKeeper)
	antehandler := sdk.ChainAnteDecorators(isd)

	testCases := []struct {
		ctx         sdk.Context
		simulate    bool
		expectedSeq uint64
	}{
		{suite.ctx.WithIsReCheckTx(true), false, 1},
		{suite.ctx.WithIsCheckTx(true).WithIsReCheckTx(false), false, 2},
		{suite.ctx.WithIsReCheckTx(true), false, 3},
		{suite.ctx.WithIsReCheckTx(true), false, 4},
		{suite.ctx.WithIsReCheckTx(true), true, 5},
	}

	for i, tc := range testCases {
		_, err := antehandler(tc.ctx, tx, tc.simulate)
		require.NoError(t, err, "unexpected error; tc #%d, %v", i, tc)
		require.Equal(t, tc.expectedSeq, suite.accountKeeper.GetAccount(suite.ctx, addr).GetSequence())
	}
}
