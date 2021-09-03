package ante_test

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

func (suite *AnteTestSuite) TestSigVerification() {
	suite.SetupTest(true) // setup
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
		acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr)
		suite.Require().NoError(acc.SetAccountNumber(uint64(i)))
		suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
		msgs[i] = testdata.NewTestMsg(addr)
	}

	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()

	spkd := ante.NewSetPubKeyDecorator(suite.app.AccountKeeper)
	svd := ante.NewSigVerificationDecorator(suite.app.AccountKeeper, suite.clientCtx.TxConfig.SignModeHandler())
	antehandler := sdk.ChainAnteDecorators(spkd, svd)

	type testCase struct {
		name      string
		privs     []cryptotypes.PrivKey
		accNums   []uint64
		accSeqs   []uint64
		recheck   bool
		shouldErr bool
	}
	testCases := []testCase{
		{"no signers", []cryptotypes.PrivKey{}, []uint64{}, []uint64{}, false, true},
		{"not enough signers", []cryptotypes.PrivKey{priv1, priv2}, []uint64{0, 1}, []uint64{0, 0}, false, true},
		{"wrong order signers", []cryptotypes.PrivKey{priv3, priv2, priv1}, []uint64{2, 1, 0}, []uint64{0, 0, 0}, false, true},
		{"wrong accnums", []cryptotypes.PrivKey{priv1, priv2, priv3}, []uint64{7, 8, 9}, []uint64{0, 0, 0}, false, true},
		{"wrong sequences", []cryptotypes.PrivKey{priv1, priv2, priv3}, []uint64{0, 1, 2}, []uint64{3, 4, 5}, false, true},
		{"valid tx", []cryptotypes.PrivKey{priv1, priv2, priv3}, []uint64{0, 1, 2}, []uint64{0, 0, 0}, false, false},
		{"no err on recheck", []cryptotypes.PrivKey{}, []uint64{}, []uint64{}, true, false},
	}
	for i, tc := range testCases {
		suite.ctx = suite.ctx.WithIsReCheckTx(tc.recheck)
		suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder() // Create new txBuilder for each test

		suite.Require().NoError(suite.txBuilder.SetMsgs(msgs...))
		suite.txBuilder.SetFeeAmount(feeAmount)
		suite.txBuilder.SetGasLimit(gasLimit)

		tx, err := suite.CreateTestTx(tc.privs, tc.accNums, tc.accSeqs, suite.ctx.ChainID())
		suite.Require().NoError(err)

		_, err = antehandler(suite.ctx, tx, false)
		if tc.shouldErr {
			suite.Require().NotNil(err, "TestCase %d: %s did not error as expected", i, tc.name)
		} else {
			suite.Require().Nil(err, "TestCase %d: %s errored unexpectedly. Err: %v", i, tc.name, err)
		}
	}
}

// This test is exactly like the one above, but we set the codec explicitly to
// Amino.
// Once https://github.com/cosmos/cosmos-sdk/issues/6190 is in, we can remove
// this, since it'll be handled by the test matrix.
// In the meantime, we want to make double-sure amino compatibility works.
// ref: https://github.com/cosmos/cosmos-sdk/issues/7229
func (suite *AnteTestSuite) TestSigVerification_ExplicitAmino() {
	suite.app, suite.ctx = createTestApp(suite.T(), true)
	suite.ctx = suite.ctx.WithBlockHeight(1)

	// Set up TxConfig.
	aminoCdc := codec.NewLegacyAmino()
	// We're using TestMsg amino encoding in some tests, so register it here.
	txConfig := legacytx.StdTxConfig{Cdc: aminoCdc}

	suite.clientCtx = client.Context{}.
		WithTxConfig(txConfig)

	anteHandler, err := ante.NewAnteHandler(
		ante.HandlerOptions{
			AccountKeeper:   suite.app.AccountKeeper,
			BankKeeper:      suite.app.BankKeeper,
			FeegrantKeeper:  suite.app.FeeGrantKeeper,
			SignModeHandler: txConfig.SignModeHandler(),
			SigGasConsumer:  ante.DefaultSigVerificationGasConsumer,
		},
	)

	suite.Require().NoError(err)
	suite.anteHandler = anteHandler

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
		acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr)
		suite.Require().NoError(acc.SetAccountNumber(uint64(i)))
		suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
		msgs[i] = testdata.NewTestMsg(addr)
	}

	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()

	spkd := ante.NewSetPubKeyDecorator(suite.app.AccountKeeper)
	svd := ante.NewSigVerificationDecorator(suite.app.AccountKeeper, suite.clientCtx.TxConfig.SignModeHandler())
	antehandler := sdk.ChainAnteDecorators(spkd, svd)

	type testCase struct {
		name      string
		privs     []cryptotypes.PrivKey
		accNums   []uint64
		accSeqs   []uint64
		recheck   bool
		shouldErr bool
	}
	testCases := []testCase{
		{"no signers", []cryptotypes.PrivKey{}, []uint64{}, []uint64{}, false, true},
		{"not enough signers", []cryptotypes.PrivKey{priv1, priv2}, []uint64{0, 1}, []uint64{0, 0}, false, true},
		{"wrong order signers", []cryptotypes.PrivKey{priv3, priv2, priv1}, []uint64{2, 1, 0}, []uint64{0, 0, 0}, false, true},
		{"wrong accnums", []cryptotypes.PrivKey{priv1, priv2, priv3}, []uint64{7, 8, 9}, []uint64{0, 0, 0}, false, true},
		{"wrong sequences", []cryptotypes.PrivKey{priv1, priv2, priv3}, []uint64{0, 1, 2}, []uint64{3, 4, 5}, false, true},
		{"valid tx", []cryptotypes.PrivKey{priv1, priv2, priv3}, []uint64{0, 1, 2}, []uint64{0, 0, 0}, false, false},
		{"no err on recheck", []cryptotypes.PrivKey{}, []uint64{}, []uint64{}, true, false},
	}
	for i, tc := range testCases {
		suite.ctx = suite.ctx.WithIsReCheckTx(tc.recheck)
		suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder() // Create new txBuilder for each test

		suite.Require().NoError(suite.txBuilder.SetMsgs(msgs...))
		suite.txBuilder.SetFeeAmount(feeAmount)
		suite.txBuilder.SetGasLimit(gasLimit)

		tx, err := suite.CreateTestTx(tc.privs, tc.accNums, tc.accSeqs, suite.ctx.ChainID())
		suite.Require().NoError(err)

		_, err = antehandler(suite.ctx, tx, false)
		if tc.shouldErr {
			suite.Require().NotNil(err, "TestCase %d: %s did not error as expected", i, tc.name)
		} else {
			suite.Require().Nil(err, "TestCase %d: %s errored unexpectedly. Err: %v", i, tc.name, err)
		}
	}
}

func (suite *AnteTestSuite) TestSigIntegration() {
	// generate private keys
	privs := []cryptotypes.PrivKey{
		secp256k1.GenPrivKey(),
		secp256k1.GenPrivKey(),
		secp256k1.GenPrivKey(),
	}

	params := types.DefaultParams()
	initialSigCost := params.SigVerifyCostSecp256k1
	initialCost, err := suite.runSigDecorators(params, false, privs...)
	suite.Require().Nil(err)

	params.SigVerifyCostSecp256k1 *= 2
	doubleCost, err := suite.runSigDecorators(params, false, privs...)
	suite.Require().Nil(err)

	suite.Require().Equal(initialSigCost*uint64(len(privs)), doubleCost-initialCost)
}

func (suite *AnteTestSuite) runSigDecorators(params types.Params, _ bool, privs ...cryptotypes.PrivKey) (sdk.Gas, error) {
	suite.SetupTest(true) // setup
	suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()

	// Make block-height non-zero to include accNum in SignBytes
	suite.ctx = suite.ctx.WithBlockHeight(1)
	suite.app.AccountKeeper.SetParams(suite.ctx, params)

	msgs := make([]sdk.Msg, len(privs))
	accNums := make([]uint64, len(privs))
	accSeqs := make([]uint64, len(privs))
	// set accounts and create msg for each address
	for i, priv := range privs {
		addr := sdk.AccAddress(priv.PubKey().Address())
		acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr)
		suite.Require().NoError(acc.SetAccountNumber(uint64(i)))
		suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
		msgs[i] = testdata.NewTestMsg(addr)
		accNums[i] = uint64(i)
		accSeqs[i] = uint64(0)
	}
	suite.Require().NoError(suite.txBuilder.SetMsgs(msgs...))

	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()
	suite.txBuilder.SetFeeAmount(feeAmount)
	suite.txBuilder.SetGasLimit(gasLimit)

	tx, err := suite.CreateTestTx(privs, accNums, accSeqs, suite.ctx.ChainID())
	suite.Require().NoError(err)

	spkd := ante.NewSetPubKeyDecorator(suite.app.AccountKeeper)
	svgc := ante.NewSigGasConsumeDecorator(suite.app.AccountKeeper, ante.DefaultSigVerificationGasConsumer)
	svd := ante.NewSigVerificationDecorator(suite.app.AccountKeeper, suite.clientCtx.TxConfig.SignModeHandler())
	antehandler := sdk.ChainAnteDecorators(spkd, svgc, svd)

	// Determine gas consumption of antehandler with default params
	before := suite.ctx.GasMeter().GasConsumed()
	ctx, err := antehandler(suite.ctx, tx, false)
	after := ctx.GasMeter().GasConsumed()

	return after - before, err
}

func (suite *AnteTestSuite) TestIncrementSequenceDecorator() {
	suite.SetupTest(true) // setup
	suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()

	priv, _, addr := testdata.KeyTestPubAddr()
	acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr)
	suite.Require().NoError(acc.SetAccountNumber(uint64(50)))
	suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

	msgs := []sdk.Msg{testdata.NewTestMsg(addr)}
	suite.Require().NoError(suite.txBuilder.SetMsgs(msgs...))
	privs := []cryptotypes.PrivKey{priv}
	accNums := []uint64{suite.app.AccountKeeper.GetAccount(suite.ctx, addr).GetAccountNumber()}
	accSeqs := []uint64{suite.app.AccountKeeper.GetAccount(suite.ctx, addr).GetSequence()}
	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()
	suite.txBuilder.SetFeeAmount(feeAmount)
	suite.txBuilder.SetGasLimit(gasLimit)

	tx, err := suite.CreateTestTx(privs, accNums, accSeqs, suite.ctx.ChainID())
	suite.Require().NoError(err)

	isd := ante.NewIncrementSequenceDecorator(suite.app.AccountKeeper)
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
		suite.Require().NoError(err, "unexpected error; tc #%d, %v", i, tc)
		suite.Require().Equal(tc.expectedSeq, suite.app.AccountKeeper.GetAccount(suite.ctx, addr).GetSequence())
	}
}
